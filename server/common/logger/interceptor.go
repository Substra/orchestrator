package logger

import (
	"context"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/server/common/trace"
	"google.golang.org/grpc"
)

// AddLogger adds a logger to the context. The logger embeds a "request id" field which is set to a random string for each context.
// Effectively, each end-user gRPC request is assigned a unique "request id".  Use this logger throughout the request lifecycle using
// `logger.Get(ctx)`. This ensures all the log entries have the same "request id" field, making it easy filter log entries by
// "request id".
func AddLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger := log.WithField("requestID", trace.GetRequestID(ctx))

	ctx = log.SetContext(ctx, logger)

	return handler(ctx, req)
}
