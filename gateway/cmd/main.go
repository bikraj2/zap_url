package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bikraj2/url_shortener/gateway/internal/controller"
	redirect "github.com/bikraj2/url_shortener/gateway/internal/gateway/redirect/http"
	shorten "github.com/bikraj2/url_shortener/gateway/internal/gateway/shorten/http"
	httphandler "github.com/bikraj2/url_shortener/gateway/internal/handler/http"
	"github.com/bikraj2/url_shortener/pkg/discovery"
	"github.com/bikraj2/url_shortener/pkg/discovery/consul"
	"github.com/go-chi/chi/v5"
)

const ServiceName = "shorten_gateway"

type config struct {
	cors struct {
		trustedOrigins []string
	}
}
type application struct {
	cfg config
}

func main() {
	var cfg config
	flag.Func("cors-trusted-origin", "Trusted CORS orirgins (space seperated)", func(s string) error {
		cfg.cors.trustedOrigins = strings.Fields(s)
		return nil
	})
	flag.Parse()
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

	app := &application{cfg: cfg}
	redirectGateway := redirect.New(registry)
	shortenGateway := shorten.New(registry)
	ctrl := controller.New(redirectGateway, shortenGateway)
	h := httphandler.New(ctrl)

	r := chi.NewRouter()
	r.Use(app.enableCORS)
	r.Use(loggingMiddleware)
	r.Get("/", h.Redirect)
	r.Get("/api/v1/resolve/{short_url}", h.GetLongUrl)
	r.Post("/api/v1/shorten", h.CreateShortUrl)
	server := &http.Server{
		Addr:    ":8084",
		Handler: r}
	fmt.Println("server is listening on port: 8084")
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")
		if origin != "" && len(app.cfg.cors.trustedOrigins) != 0 {
			for i := range app.cfg.cors.trustedOrigins {
				if origin == app.cfg.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Request-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authentication, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the request details (method, URL, and time)
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		// Create a custom ResponseWriter to capture status code and log it
		lrw := &loggedResponseWriter{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log the response status code and the time taken to process the request
		log.Printf("Completed %s %s with status %d in %v", r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
	})
}

// loggedResponseWriter is a wrapper around http.ResponseWriter to capture status code
type loggedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code from the response
func (lrw *loggedResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}
