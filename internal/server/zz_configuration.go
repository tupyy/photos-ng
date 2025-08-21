package server

import (
	"time"

	defaults "github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
)

type HttpServerConfigOption func(r *HttpServerConfig)

// NewHttpServerConfigWithOptions creates a new HttpServerConfig with the passed in options set
func NewHttpServerConfigWithOptions(opts ...HttpServerConfigOption) *HttpServerConfig {
	r := &HttpServerConfig{}
	for _, o := range opts {
		o(r)
	}
	return r
}

// NewHttpServerConfigWithOptionsAndDefaults creates a new HttpServerConfig with the passed in options set starting from the defaults
func NewHttpServerConfigWithOptionsAndDefaults(opts ...HttpServerConfigOption) *HttpServerConfig {
	r := &HttpServerConfig{
		Port:               8080,
		RegisterHandlersFn: make(map[string]func(router *gin.RouterGroup)),
	}
	defaults.MustSet(r)
	for _, o := range opts {
		o(r)
	}
	return r
}

// ToOption returns a new HttpServerConfigOption that sets the values from the passed in HttpServerConfig
func (r *HttpServerConfig) ToOption() HttpServerConfigOption {
	return func(to *HttpServerConfig) {
		to.GraceTimeout = r.GraceTimeout
		to.Port = r.Port
		to.RegisterHandlersFn = r.RegisterHandlersFn
	}
}

// HttpServerConfigWithOptions configures an existing HttpServerConfig with the passed in options set
func HttpServerConfigWithOptions(r *HttpServerConfig, opts ...HttpServerConfigOption) *HttpServerConfig {
	for _, o := range opts {
		o(r)
	}
	return r
}

// WithOptions configures the receiver HttpServerConfig with the passed in options set
func (r *HttpServerConfig) WithOptions(opts ...HttpServerConfigOption) *HttpServerConfig {
	for _, o := range opts {
		o(r)
	}
	return r
}

// WithGraceTimeout returns an option that can set GraceTimeout on a HttpServerConfig
func WithGraceTimeout(graceTimeout time.Duration) HttpServerConfigOption {
	return func(r *HttpServerConfig) {
		r.GraceTimeout = graceTimeout
	}
}

func WithRegisterHandlers(apiVersion string, fn func(engine *gin.RouterGroup)) HttpServerConfigOption {
	return func(r *HttpServerConfig) {
		r.RegisterHandlersFn[apiVersion] = fn
	}
}

func WithGinMode(ginMode string) HttpServerConfigOption {
	return func(r *HttpServerConfig) {
		r.GinMode = ginMode
	}
}

// WithHttpPort returns an option that can set Port on a HttpServerConfig
func WithHttpPort(port int) HttpServerConfigOption {
	return func(r *HttpServerConfig) {
		r.Port = port
	}
}

func WithApiVersion(apiVersion string) HttpServerConfigOption {
	return func(r *HttpServerConfig) {
		r.ApiVersion = apiVersion
	}
}

func WithMode(mode string) HttpServerConfigOption {
	return func(r *HttpServerConfig) {
		r.Mode = mode
	}
}

func WithStaticsFolder(staticsFolder string) HttpServerConfigOption {
	return func(r *HttpServerConfig) {
		r.StaticsFolder = staticsFolder
	}
}
