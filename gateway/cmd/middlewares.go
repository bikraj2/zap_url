package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//	func (app *application) enableCORS(c *gin.Context) {
//		origin := c.GetHeader("Origin")
//		c.Writer.Header().Set("Vary", "Origin")
//		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
//		c.Writer.Header().Set("Vary", "Access-Control-Request-Method")
//
//		if origin != "" && len(app.cfg.Cors.TrustedOrigins) != 0 {
//			for _, allowedOrigin := range app.cfg.Cors.TrustedOrigins {
//				if origin == allowedOrigin {
//					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
//					if c.Request.Method == http.MethodOptions {
//						c.Writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
//						c.Writer.Header().Set("Access-Control-Allow-Headers", "Authentication, Content-Type")
//						c.AbortWithStatus(http.StatusOK)
//						return
//					}
//				}
//			}
//		}
//		c.Next()
//	}
func (app *application) loggingMiddleware(c *gin.Context) {
	start := time.Now()

	// Process the request
	c.Next()

	// Log the request details in a structured way
	app.logger.Info("Request completed",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("status", c.Writer.Status()),
		zap.String("client_ip", c.ClientIP()),
		zap.Duration("latency", time.Since(start)),
	)
}
