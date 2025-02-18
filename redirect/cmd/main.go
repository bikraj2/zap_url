package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/bikraj2/url_shortener/pkg/discovery"
	"github.com/bikraj2/url_shortener/pkg/discovery/consul"
	"github.com/bikraj2/url_shortener/redirect/internal/controller"
	httphandler "github.com/bikraj2/url_shortener/redirect/internal/hanlder/http"
	"github.com/bikraj2/url_shortener/redirect/internal/repository/postgresql"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)

type config struct {
	port int
	db   struct {
		dsn          string
		MaxOpenCons  int
		MaxIdleConns int
		MaxIdleTime  string
	}
}

const ServiceName = "redirect"

func main() {
	var cfg config

	flag.IntVar(&cfg.db.MaxOpenCons, "maxOpenCons", 25, "Maximum Number of Open Connections")
	flag.IntVar(&cfg.db.MaxIdleConns, "maxIdleCons", 23, "Maximum Number of Open Idle Connections")
	flag.StringVar(&cfg.db.MaxIdleTime, "maxIdleTime", "15m", "Maximum Number of Open Idle Connections")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://admin:admin@localhost:5432/url_shortener?sslmode=disable", "Postgresql DSN")

	registry, err := consul.New("localhost:8500")
	if err != nil {
		panic(err)

	}
	instanceID := discovery.GenerateInstanceID(ServiceName)
	err = registry.RegisterService(context.Background(), ServiceName, instanceID, "localhost:8082")
	if err != nil {
		panic(err)
	}
	defer registry.DeRegisterService(context.Background(), "", instanceID)
	go func() {
		for {
			registry.ReportHealthyState(instanceID, "")
			time.Sleep(4 * time.Second)
		}
	}()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	flag.Parse()
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	db, err := openDB(cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer db.Close()
	repo := repository.New(db, client)

	fmt.Println("Repository initialized successfully!")
	ctrl := controller.New(repo)

	fmt.Println("controller initialized successfully!")

	h := httphandler.New(ctrl)
	// Set up HTTP server
	r := mux.NewRouter()
	r.HandleFunc("/{short_url}", h.Handle).Methods(http.MethodGet)
	server := &http.Server{
		Addr:    ":8082",
		Handler: r,
	}

	fmt.Println("Server is running on http://localhost:8082")

	log.Fatal(server.ListenAndServe())
}
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.MaxOpenCons)
	db.SetMaxIdleConns(cfg.db.MaxIdleConns) // This should be lesser than  the MaxOpenCons
	duration, err := time.ParseDuration(cfg.db.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
