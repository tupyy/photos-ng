package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	keyfunc "github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	userCtx "git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/user"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

const (
	alg               = "RS256"
	sessionKey string = "sessionKey"
	stateKey   string = "state"
	tokenIDKey string = "tokenID"
	userKey    string = "user"
)

type userClaims struct {
	Username  string `json:"preferred_username"`
	Email     string `json:"email"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	jwt.RegisteredClaims
}

type OIDCAuthenticator struct {
	wellknownData map[string]interface{}
	keyFunc       keyfunc.Keyfunc
}

func NewAuthenticator(wellknownURL string) (*OIDCAuthenticator, error) {
	tracer := logger.New("auth").Operation("initialize_authenticator").
		WithString("wellknown_url", wellknownURL).
		Build()

	data, err := getWellknownData(wellknownURL)
	if err != nil {
		return nil, fmt.Errorf("invalid oidc configuration: %w", err)
	}

	tracer.Step("wellknown_fetched").
		WithString("auth_url", data["authorization_endpoint"].(string)).
		WithString("token_url", data["token_endpoint"].(string)).
		Log()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	k, err := keyfunc.NewDefaultCtx(ctx, []string{data["jwks_uri"].(string)})
	if err != nil {
		return nil, fmt.Errorf("failed to create a keyfunc.Keyfunc from the server's URL.\nError: %s", err)
	}

	tracer.Success().Log()

	return &OIDCAuthenticator{
		wellknownData: data,
		keyFunc:       k,
	}, nil
}

func (o *OIDCAuthenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		debug := logger.New("auth_middleware").WithContext(ctx)
		tracer := debug.Operation("auth_middleware").Build()

		var bearerToken string
		if cookie, err := c.Cookie("BearerToken"); err == nil {
			bearerToken = cookie
		}

		if bearerToken == "" {
			tracer.Step("bearer token not found in cookie").Log()
			c.Next()
		}
		tracer.Step("validating_token").Log()
		user, err := o.Authenticate(ctx, bearerToken)
		if err == nil {
			uCtx := userCtx.ToContext(c.Request.Context(), user)
			c.Request = c.Request.WithContext(uCtx)
			c.Set(userKey, user)

			tracer.Step("token validated").Log()
			c.Next()

			return
		}
	}
}

func (o *OIDCAuthenticator) Authenticate(ctx context.Context, token string) (*entity.User, error) {
	tracer := logger.New("auth_middleware").WithContext(ctx).Operation("authenticate").Build()

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{alg}),
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
	)

	tt, err := parser.ParseWithClaims(token, &userClaims{}, o.keyFunc.Keyfunc)
	if err != nil {
		tracer.Step("parse_with_claims").WithString("error", err.Error()).Log()
		return nil, err
	}

	claims, ok := tt.Claims.(*userClaims)
	if !ok || !tt.Valid {
		tracer.Step("invalid_claims").Log()
		return nil, fmt.Errorf("invalid token claims")
	}

	user := newUser(claims)

	tracer.Success().
		WithString("user_id", user.ID).
		WithString("username", user.Username).
		Log()

	return user, nil
}

func getWellknownData(wellknownURL string) (map[string]any, error) {
	resp, err := http.Get(wellknownURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to retrieve wellknown json: %s", resp.Status)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	wellknown := make(map[string]any)
	if err := json.Unmarshal(data, &wellknown); err != nil {
		return nil, fmt.Errorf("unable to unmarshal wellknown json: %w", err)
	}

	return wellknown, nil
}

func newUser(claims *userClaims) *entity.User {
	return &entity.User{
		ID:        claims.Subject,
		Username:  claims.Username,
		FirstName: claims.FirstName,
		LastName:  claims.LastName,
	}
}
