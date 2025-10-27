package server

import (
	"context"
	"fmt"
	"net"
	"time"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/grpc"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/server/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServerConfig struct {
	Handler      v1.PhotosNGServiceServer
	GraceTimeout time.Duration
	Port         int
	Mode         string
}

type GrpcServer struct {
	port int
	srv  *grpc.Server
}

func NewGrpcServer(cfg *GrpcServerConfig) (*GrpcServer, error) {
	// Create gRPC server with optional logging middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.LoggingInterceptor()),
		grpc.StreamInterceptor(interceptors.StreamLoggingInterceptor()),
	)

	// Create gRPC handler
	v1.RegisterPhotosNGServiceServer(grpcServer, cfg.Handler)

	// Enable gRPC reflection for development
	if cfg.Mode == "dev" {
		reflection.Register(grpcServer)
		zap.S().Info("gRPC reflection enabled for development")
	}

	return &GrpcServer{srv: grpcServer, port: cfg.Port}, nil
}

func (g *GrpcServer) Start(ctx context.Context) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port %d: %w", g.port, err)
	}

	zap.S().Named("grpc").Infof("gRPC server listening on :%d", g.port)
	if err := g.srv.Serve(listen); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}

func (g *GrpcServer) Stop(ctx context.Context) error {
	zap.S().Named("grpc").Info("server shutting down...")
	g.srv.GracefulStop()
	return nil
}
