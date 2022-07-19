package common

import (
	"context"

	"google.golang.org/grpc"
)

type ContextualizedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *ContextualizedServerStream) Context() context.Context {
	return w.ctx
}

func BindStreamToContext(ctx context.Context, stream grpc.ServerStream) *ContextualizedServerStream {
	return &ContextualizedServerStream{stream, ctx}
}
