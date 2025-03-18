package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bikraj2/url_shortener/gateway/internal/controller"
	redirect "github.com/bikraj2/url_shortener/gateway/internal/gateway/redirect/http"
	shorten "github.com/bikraj2/url_shortener/gateway/internal/gateway/shorten/http"
	httphandler "github.com/bikraj2/url_shortener/gateway/internal/handler/http"
	"github.com/bikraj2/url_shortener/gateway/internal/repository/data"
	"github.com/bikraj2/url_shortener/pkg/discovery"
	"github.com/bikraj2/url_shortener/pkg/discovery/consul"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const ServiceName = "shorten_gateway"

type config struct {
	Cors struct {
		TrustedOrigins []string `yaml:"trusted_origins"`
	} `yaml:"cors"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Consul struct {
		Address string `yaml:"address"`
	} `yaml:"consul"`
	DB struct {
		Dsn          string `yaml:"dsn"`
		MaxOpenCons  int    `yaml:"max_open_cons"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
		MaxIdleTime  string `yaml:"max_idle_time"`
	} `yaml:"db"`
}
type application struct {
	cfg    config `yaml:"cfg"`
	logger *zap.Logger
	db     *sql.DB
}

func main() {

	zap_cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      true,
		Encoding:         "json",
		OutputPaths:      []string{"stdout", "./log/app.log"},
		ErrorOutputPaths: []string{"stderr", "./log/error.log"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "timestamp",
			LevelKey:     "level",
			MessageKey:   "msg",
			CallerKey:    "caller",
			EncodeTime:   zapcore.ISO8601TimeEncoder,  // Human-readable timestamps
			EncodeLevel:  zapcore.CapitalLevelEncoder, // INFO, ERROR, etc.
			EncodeCaller: zapcore.ShortCallerEncoder,  // filename:line_number
		},
	}
	logger, err := zap_cfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg config
	err = loadConfig(&cfg)
	if err != nil {
		panic(err)
	}
	registry, err := consul.New("dev-consul:8500")
	if err != nil {
		panic(err)
	}
	instanceID := discovery.GenerateInstanceID(ServiceName)
	err = registry.RegisterService(context.Background(), ServiceName, instanceID, "gateway:8084")
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

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal("Error while connecting to the database", zap.String("error", err.Error()))
	}
	app := &application{cfg: cfg, logger: logger, db: db}

	redirectGateway := redirect.New(registry)
	shortenGateway := shorten.New(registry)
	userRepo := data.New(app.db)
	ctrl := controller.New(redirectGateway, shortenGateway, userRepo)
	h := httphandler.New(ctrl, app.logger)

	r := gin.New()

	r.SetTrustedProxies(app.cfg.Cors.TrustedOrigins)
	r.Use(app.loggingMiddleware)
	r.GET("/", h.Redirect)
	r.GET("/api/v1/resolve/:short_url", h.GetLongUrl)
	r.POST("/api/v1/shorten", h.CreateShortUrl)
	err = r.Run(fmt.Sprintf(":%s", app.cfg.Server.Port))
	if err != nil {
		logger.Fatal("Error while Starting the sever", zap.String("error", err.Error()))
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.Dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.DB.MaxOpenCons)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns) // This should be lesser than  the MaxOpenCons
	duration, err := time.ParseDuration(cfg.DB.MaxIdleTime)
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
