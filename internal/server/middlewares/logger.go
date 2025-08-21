package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger returns a gin middleware that logs HTTP requests using zap logger.
// It logs the method, path, status code, and response time for each request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		track := true

		if track {
			end := time.Now()
			latency := end.Sub(start)

			fields := []zapcore.Field{
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Duration("latency", latency),
				zap.String("time", end.Format(time.RFC3339)),
			}

			if len(c.Errors) > 0 {
				// Append error field if this is an erroneous request.
				for _, e := range c.Errors.Errors() {
					zap.S().Named("http").Desugar().Error(e, fields...)
				}
			} else {
				zap.S().Named("http").Desugar().Info(path, fields...)
			}
		}
	}
}
