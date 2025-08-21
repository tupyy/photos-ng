package server

import (
	"time"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/grpc"
	defaults "github.com/creasty/defaults"
)

type GrpcServerConfigOption func(r *GrpcServerConfig)

// NewGrpcServerConfigWithOptions creates a new GrpcServerConfig with the passed in options set
func NewGrpcServerConfigWithOptions(opts ...GrpcServerConfigOption) *GrpcServerConfig {
	r := &GrpcServerConfig{}
	for _, o := range opts {
		o(r)
	}
	return r
}

// NewGrpcServerConfigWithOptionsAndDefaults creates a new GrpcServerConfig with the passed in options set starting from the defaults
func NewGrpcServerConfigWithOptionsAndDefaults(opts ...GrpcServerConfigOption) *GrpcServerConfig {
	r := &GrpcServerConfig{
		Port: 9090,
	}
	defaults.MustSet(r)
	for _, o := range opts {
		o(r)
	}
	return r
}

// ToOption returns a new GrpcServerConfigOption that sets the values from the passed in GrpcServerConfig
func (r *GrpcServerConfig) ToOption() GrpcServerConfigOption {
	return func(to *GrpcServerConfig) {
		to.Handler = r.Handler
		to.GraceTimeout = r.GraceTimeout
		to.Port = r.Port
		to.Mode = r.Mode
	}
}

// GrpcServerConfigWithOptions configures an existing GrpcServerConfig with the passed in options set
func GrpcServerConfigWithOptions(r *GrpcServerConfig, opts ...GrpcServerConfigOption) *GrpcServerConfig {
	for _, o := range opts {
		o(r)
	}
	return r
}

// WithOptions configures the receiver GrpcServerConfig with the passed in options set
func (r *GrpcServerConfig) WithOptions(opts ...GrpcServerConfigOption) *GrpcServerConfig {
	for _, o := range opts {
		o(r)
	}
	return r
}

// WithHandler returns an option that can set Handler on a GrpcServerConfig
func WithGrpcHandler(handler v1.PhotosNGServiceServer) GrpcServerConfigOption {
	return func(r *GrpcServerConfig) {
		r.Handler = handler
	}
}

// WithGrpcGraceTimeout returns an option that can set GraceTimeout on a GrpcServerConfig
func WithGrpcGraceTimeout(graceTimeout time.Duration) GrpcServerConfigOption {
	return func(r *GrpcServerConfig) {
		r.GraceTimeout = graceTimeout
	}
}

// WithGrpcPort returns an option that can set Port on a GrpcServerConfig
func WithGrpcPort(port int) GrpcServerConfigOption {
	return func(r *GrpcServerConfig) {
		r.Port = port
	}
}

// WithGrpcMode returns an option that can set Mode on a GrpcServerConfig
func WithGrpcMode(mode string) GrpcServerConfigOption {
	return func(r *GrpcServerConfig) {
		r.Mode = mode
	}
}
