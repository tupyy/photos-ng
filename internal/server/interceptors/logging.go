package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor is a gRPC interceptor for unary request logging using zap
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Get peer info for client address
		var clientAddr string
		if p, ok := peer.FromContext(ctx); ok {
			clientAddr = p.Addr.String()
		}

		startLogger := zap.S().Named("grpc").With(
			"method", info.FullMethod,
			"client_addr", clientAddr,
			"time", start.Format(time.RFC3339),
		)

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if requestId, ok := md["x-request-id"]; ok {
				startLogger.With("request_id", requestId)
			}
		}

		startLogger.Info("gRPC unary call started")

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		st := status.Code(err)

		// Create logger with common fields
		logger := zap.S().Named("grpc").With(
			"method", info.FullMethod,
			"duration_ms", float64(duration.Nanoseconds())/1e6,
			"status_code", st.String(),
			"client_addr", clientAddr,
		)

		// Log based on status
		if err != nil {
			if st == codes.Internal || st == codes.Unknown {
				logger.Errorw("gRPC unary call failed", "error", err)
			} else {
				logger.Warnw("gRPC unary call completed with error", "error", err)
			}
		} else {
			logger.Infow("gRPC unary call completed successfully")
		}

		return resp, err
	}
}

// StreamLoggingInterceptor is a gRPC interceptor for streaming request logging using zap
func StreamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Get peer info for client address
		var clientAddr string
		if p, ok := peer.FromContext(ss.Context()); ok {
			clientAddr = p.Addr.String()
		}

		startLogger := zap.S().Named("grpc").With(
			"method", info.FullMethod,
			"client_addr", clientAddr,
			"time", start.Format(time.RFC3339),
		)

		if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
			if requestId, ok := md["x-request-id"]; ok {
				startLogger.With("request_id", requestId)
			}
		}

		startLogger.Info("gRPC stream call started")

		// Call the handler
		err := handler(srv, ss)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		st := status.Code(err)

		// Create logger with common fields
		logger := zap.S().Named("grpc").With(
			"method", info.FullMethod,
			"duration_ms", float64(duration.Nanoseconds())/1e6,
			"status_code", st.String(),
			"client_addr", clientAddr,
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		)

		// Log based on status
		if err != nil {
			if st == codes.Internal || st == codes.Unknown {
				logger.Errorw("gRPC stream call failed", "error", err)
			} else {
				logger.Warnw("gRPC stream call completed with error", "error", err)
			}
		} else {
			logger.Infow("gRPC stream call completed successfully")
		}

		return err
	}
}
