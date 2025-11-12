package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type contextKey string

const userKey contextKey = "user"

// Generate creates a new unique request ID
func Generate() string {
	return uuid.New().String()
}

// ToContext adds a request ID to the context
func ToContext(ctx context.Context, user *entity.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// FromContext extracts the request ID from the context.
// Returns empty string if request ID is not found.
func FromContext(ctx context.Context) *entity.User {
	if user, ok := ctx.Value(userKey).(*entity.User); ok {
		return user
	}
	return nil
}

// FromGin extracts the request ID from gin.Context.
// Returns empty string if request ID is not found.
func FromGin(c *gin.Context) *entity.User {
	if userV, ok := c.Get(string(userKey)); ok {
		if user, ok := userV.(*entity.User); ok {
			return user
		}
	}
	return nil
}
