package middlewares

import (
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/requestid"
	"github.com/gin-gonic/gin"
)

// RequestID gets the request ID from Envoy's x-request-id header or generates
// a unique request ID for each HTTP request and injects it into both the
// gin.Context and the request's context.Context for consistent access across
// the application layer.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First try to get request ID from Envoy header
		requestID := c.GetHeader("x-request-id")

		// If no header provided, generate a new UUID
		if requestID == "" {
			requestID = requestid.Generate()
		}

		// Store in gin.Context for easy access in handlers
		c.Set(string(requestid.RequestIDKey), requestID)

		// Create new context with requestID and replace request context
		ctx := requestid.ToContext(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetRequestIDFromGin extracts the request ID from gin.Context.
// Returns empty string if request ID is not found.
func GetRequestIDFromGin(c *gin.Context) string {
	return requestid.FromGin(c)
}
