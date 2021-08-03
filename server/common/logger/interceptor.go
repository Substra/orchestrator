package logger

import (
	"context"
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// AddLogger adds a logger to the context. The logger embeds a "request id" field which is set to a random string for each context.
// Effectively, each end-user gRPC request is assigned a unique "request id".  Use this logger throughout the request lifecycle using
// `logger.Get(ctx)`. This ensures all the log entries have the same "request id" field, making it easy filter log entries by
// "request id".
func AddLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	logger := log.WithField("request id", fmt.Sprintf("%v", u)[:8])

	ctx = log.SetContext(ctx, logger)

	return handler(ctx, req)
}
