package common

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.Info("request",
			slog.M{"method": c.Request.Method,
				"path":      c.Request.URL.Path,
				"status":    c.Writer.Status(),
				"latency":   latency,
				"client_ip": c.ClientIP(),
			},
		)
	}
}
