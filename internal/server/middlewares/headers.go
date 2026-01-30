package middlewares

import "github.com/gin-gonic/gin"

// Headers returns a gin middleware that sets common HTTP headers for all responses.
func Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict browser features
		c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()")

		// Cross-origin isolation headers
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")

		// Cache control for sensitive data
		c.Header("Cache-Control", "no-store")

		c.Next()
	}
}
