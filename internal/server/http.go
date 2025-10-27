package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/server/middlewares"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/requestid"
)

const (
	ProductionServer string = "prod"
	DevServer        string = "dev"
)

type HttpServerConfig struct {
	GraceTimeout       time.Duration
	Port               int
	RegisterHandlersFn map[string]func(router *gin.RouterGroup)
	GinMode            string
	ApiVersion         string
	Mode               string
	StaticsFolder      string
}

type HttpServer struct {
	srv    *http.Server
	cfg    *HttpServerConfig
	engine *gin.Engine
}

// NewHttpServer creates a new runnable server instance with the provided configuration.
func NewHttpServer(cfg *HttpServerConfig) *HttpServer {
	gin.SetMode(cfg.GinMode)
	engine := gin.New()

	if cfg.Mode == ProductionServer {
		// Serve static files from ui/dist directory (for frontend)
		engine.Static("/static", cfg.StaticsFolder)
		engine.StaticFile("/", path.Join(cfg.StaticsFolder, "index.html"))
		engine.StaticFile("/favicon.ico", path.Join(cfg.StaticsFolder, "favicon.ico"))

		// Serve index.html for any non-API routes (SPA routing support)
		engine.NoRoute(func(c *gin.Context) {
			// Don't serve index.html for API routes
			if c.Request.URL.Path[:4] == "/api" {
				requestID := requestid.FromGin(c)
				c.JSON(404, gin.H{
					"error": "API endpoint not found",
					"request_id": requestID,
				})
				return
			}
			c.File(path.Join(cfg.StaticsFolder, "index.html"))
		})
	}

	// for each api version register handlers
	for apiVersion, registerHandlersFn := range cfg.RegisterHandlersFn {
		router := engine.Group(apiVersion)
		router.Use(
			middlewares.Headers(),
			middlewares.RequestID(),
			middlewares.Logger(),
			ginzap.RecoveryWithZap(zap.S().Desugar(), true),
		)
		registerHandlersFn(router)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler: engine,
	}

	return &HttpServer{srv: srv, cfg: cfg}
}

// Run starts the HTTP server and handles graceful shutdown when the context is cancelled.
func (r *HttpServer) Start(ctx context.Context) error {
	if err := r.srv.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		zap.S().Named("http").Errorw("failed to start server", "error", err)
		return err
	}

	return nil
}

func (r *HttpServer) Stop(ctx context.Context) error {
	zap.S().Named("http").Info("server shutting down...")
	return r.srv.Shutdown(ctx)
}
