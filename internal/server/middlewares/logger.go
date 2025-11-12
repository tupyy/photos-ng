package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/user"
)

// Logger returns a gin middleware that logs HTTP requests using zap logger.
// It logs request start with requestId and all fields except status, then request end with requestId and status.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		requestID := GetRequestIDFromGin(c)
		usr := user.FromGin(c)

		// Log request start with requestId and current fields (except status)
		startFields := []zapcore.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("time", start.Format(time.RFC3339)),
		}

		if usr != nil {
			startFields = append(startFields, zap.String("username", usr.Username))
		}

		zap.S().Named("http").Desugar().Info("Request started", startFields...)

		c.Next()

		// Log request end with requestId and status
		end := time.Now()
		latency := end.Sub(start)

		endFields := []zapcore.Field{
			zap.String("request_id", requestID),
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.String("time", end.Format(time.RFC3339)),
		}

		if usr != nil {
			endFields = append(endFields, zap.String("username", usr.Username))
		}

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				zap.S().Named("http").Desugar().Error(e, endFields...)
			}
		} else {
			zap.S().Named("http").Desugar().Info("Request completed", endFields...)
		}
	}
}
