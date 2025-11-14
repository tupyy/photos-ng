package user

import (
	"context"

	"github.com/gin-gonic/gin"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type contextKey string

const userKey contextKey = "user"

func ToContext(ctx context.Context, user *entity.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func MustFromContext(ctx context.Context) entity.User {
	user := FromContext(ctx)
	if user == nil {
		panic("user not found in context")
	}
	return *user
}

func FromContext(ctx context.Context) *entity.User {
	if user, ok := ctx.Value(userKey).(*entity.User); ok {
		return user
	}
	return nil
}

func FromGin(c *gin.Context) *entity.User {
	if userV, ok := c.Get(string(userKey)); ok {
		if user, ok := userV.(*entity.User); ok {
			return user
		}
	}
	return nil
}
