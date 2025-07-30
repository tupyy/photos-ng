package middlewares

import "github.com/gin-gonic/gin"

// Headers returns a gin middleware that sets common HTTP headers for all responses.
// Currently sets the Content-Type header to application/json.
func Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	}
}
