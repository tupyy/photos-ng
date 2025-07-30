package server

import (
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	defaults "github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
)

type RunnableServerConfigOption func(r *RunnableServerConfig)

// NewRunnableServerConfigWithOptions creates a new RunnableServerConfig with the passed in options set
func NewRunnableServerConfigWithOptions(opts ...RunnableServerConfigOption) *RunnableServerConfig {
	r := &RunnableServerConfig{}
	for _, o := range opts {
		o(r)
	}
	return r
}

// NewRunnableServerConfigWithOptionsAndDefaults creates a new RunnableServerConfig with the passed in options set starting from the defaults
func NewRunnableServerConfigWithOptionsAndDefaults(opts ...RunnableServerConfigOption) *RunnableServerConfig {
	r := &RunnableServerConfig{
		Port:               8080,
		RegisterHandlersFn: make(map[string]func(router *gin.RouterGroup)),
	}
	defaults.MustSet(r)
	for _, o := range opts {
		o(r)
	}
	return r
}

// ToOption returns a new RunnableServerConfigOption that sets the values from the passed in RunnableServerConfig
func (r *RunnableServerConfig) ToOption() RunnableServerConfigOption {
	return func(to *RunnableServerConfig) {
		to.GraceTimeout = r.GraceTimeout
		to.Port = r.Port
		to.RegisterHandlersFn = r.RegisterHandlersFn
	}
}

// RunnableServerConfigWithOptions configures an existing RunnableServerConfig with the passed in options set
func RunnableServerConfigWithOptions(r *RunnableServerConfig, opts ...RunnableServerConfigOption) *RunnableServerConfig {
	for _, o := range opts {
		o(r)
	}
	return r
}

// WithOptions configures the receiver RunnableServerConfig with the passed in options set
func (r *RunnableServerConfig) WithOptions(opts ...RunnableServerConfigOption) *RunnableServerConfig {
	for _, o := range opts {
		o(r)
	}
	return r
}

// WithGraceTimeout returns an option that can set GraceTimeout on a RunnableServerConfig
func WithGraceTimeout(graceTimeout time.Duration) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.GraceTimeout = graceTimeout
	}
}

func WithRegisterHandlers(apiVersion string, fn func(engine *gin.RouterGroup)) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.RegisterHandlersFn[apiVersion] = fn
	}
}

func WithDatastore(dt *pg.Datastore) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.Datastore = dt
	}
}

func WithCloseCallback(cb func() error) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.CloseCb = cb
	}
}

func WithGinMode(ginMode string) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.GinMode = ginMode
	}
}

// WithPort returns an option that can set Port on a RunnableServerConfig
func WithPort(port int) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.Port = port
	}
}

func WithApiVersion(apiVersion string) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.ApiVersion = apiVersion
	}
}

func WithMode(mode string) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.Mode = mode
	}
}

func WithStaticsFolder(staticsFolder string) RunnableServerConfigOption {
	return func(r *RunnableServerConfig) {
		r.StaticsFolder = staticsFolder
	}
}
