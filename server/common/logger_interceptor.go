package common

import (
	"context"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/server/common/trace"
	"google.golang.org/grpc"
)

// setContext adds a logger to the context. The logger embeds a "request id" field which is set to a random string for each context.
// Effectively, each end-user gRPC request is assigned a unique "request id". Use this logger throughout the request lifecycle using
// `logger.Get(ctx)`. This ensures all the log entries have the same "request id" field, making it easy filter log entries by
// "request id".
func setContext(ctx context.Context) context.Context {
	logger := log.WithField("requestID", trace.GetRequestID(ctx))
	return log.SetContext(ctx, logger)
}

// UnaryServerLoggerInterceptor adds a logger to the context.
func UnaryServerLoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = setContext(ctx)
	return handler(ctx, req)
}

// StreamServerLoggerInterceptor adds a logger to the context.
func StreamServerLoggerInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := setContext(stream.Context())
	streamWithContext := BindStreamToContext(ctx, stream)
	return handler(srv, streamWithContext)
}
