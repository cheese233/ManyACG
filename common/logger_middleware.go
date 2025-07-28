package common

import (
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

func GinSlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.Trace("request",
			slog.M{"method": c.Request.Method,
				"path":      c.Request.URL.Path,
				"status":    c.Writer.Status(),
				"latency":   latency,
				"client_ip": c.ClientIP(),
			},
		)
	}
}

func VipsLogger(messageDomain string, verbosity vips.LogLevel, message string) {
	var logLevel slog.Level
	switch verbosity {
	case vips.LogLevelError:
		logLevel = slog.ErrorLevel
	case vips.LogLevelCritical:
		logLevel = slog.FatalLevel
	case vips.LogLevelWarning:
		logLevel = slog.WarnLevel
	case vips.LogLevelMessage:
		logLevel = slog.InfoLevel
	case vips.LogLevelInfo:
		logLevel = slog.DebugLevel
	case vips.LogLevelDebug:
		logLevel = slog.TraceLevel
	}

	Logger.Logf(logLevel, "[%v.%v] %v", messageDomain, verbosity, message)
}
