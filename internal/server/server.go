package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/server/middlewares"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ProductionServer string = "prod"
	DevServer        string = "dev"
)

type Server interface {
	Run(context.Context)
}

type RunnableServerConfig struct {
	Datastore    *pg.Datastore
	GraceTimeout time.Duration
	Port         int
	// RegisterHandlersFn holds a map of handler for each api version
	RegisterHandlersFn map[string]func(router *gin.RouterGroup)
	CloseCb            func() error
	GinMode            string
	ApiVersion         string
	Mode               string
	StaticsFolder      string
}

type runnableServer struct {
	srv         *http.Server
	cfg         *RunnableServerConfig
	engine      *gin.Engine
	closePostCb func() error
}

// NewRunnableServer creates a new runnable server instance with the provided configuration.
func NewRunnableServer(cfg *RunnableServerConfig) Server {
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
				c.JSON(404, gin.H{"error": "API endpoint not found"})
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
			middlewares.Logger(),
			ginzap.RecoveryWithZap(zap.S().Desugar(), true),
		)
		registerHandlersFn(router)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler: engine,
	}

	return &runnableServer{srv: srv, cfg: cfg, closePostCb: cfg.CloseCb}
}

// Run starts the HTTP server and handles graceful shutdown when the context is cancelled.
func (r *runnableServer) Run(ctx context.Context) {
	go func() {
		if err := r.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalw("server closed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Infow("shutdown server", "grace timeout", fmt.Sprintf("%s", r.cfg.GraceTimeout))

	newCtx, cancel := context.WithTimeout(ctx, r.cfg.GraceTimeout)
	defer cancel()
	go func() {
		if err := r.srv.Shutdown(newCtx); err != nil {
			zap.S().Errorw("server shutdown", "error", err)
		}
	}()

	<-newCtx.Done()
	zap.S().Info("server exiting")

	_ = r.closePostCb()
}
