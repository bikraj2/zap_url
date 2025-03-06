package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bikraj2/url_shortener/gateway/internal/controller"
	redirect "github.com/bikraj2/url_shortener/gateway/internal/gateway/redirect/http"
	shorten "github.com/bikraj2/url_shortener/gateway/internal/gateway/shorten/http"
	httphandler "github.com/bikraj2/url_shortener/gateway/internal/handler/http"
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
		Adddress string `yaml:"adddress"`
	} `yaml:"consul"`
}
type application struct {
	cfg    config `yaml:"cfg"`
	logger *zap.Logger
}

func main() {

	zap_cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      true,
		Encoding:         "json",
		OutputPaths:      []string{"stdout", "./app/log/app.log"},
		ErrorOutputPaths: []string{"stderr", "./app/log/error.log"},
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

	app := &application{cfg: cfg, logger: logger}
	redirectGateway := redirect.New(registry)
	shortenGateway := shorten.New(registry)
	ctrl := controller.New(redirectGateway, shortenGateway)
	h := httphandler.New(ctrl)

	r := gin.New()

	r.Use(app.enableCORS)
	r.Use(app.loggingMiddleware)
	r.GET("/", h.Redirect)
	r.GET("/api/v1/resolve/{short_url}", h.GetLongUrl)
	r.POST("/api/v1/shorten", h.CreateShortUrl)
	err = r.Run(fmt.Sprintf(":%s", app.cfg.Server.Port))
	if err != nil {
		logger.Fatal("Error while Starting the sever", zap.String("error", err.Error()))
	}
}
