package interceptors

import (
	"context"
	"strings"

	"errors"
	"github.com/owkin/orchestrator/server/common/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerRequestLogger log every gRPC response
func UnaryServerRequestLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range IgnoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	log := logger.Get(ctx).WithField("method", info.FullMethod)

	resp, err := handler(ctx, req)

	if err == nil {
		// Error is already logged by the error interceptor
		log.WithField("response", resp).Debug("Success response")
	}

	return resp, err
}

// StreamServerRequestLogger logs gRPC responses
func StreamServerRequestLogger(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := handler(srv, stream)
	if err == nil {
		return nil
	}

	log := logger.
		Get(stream.Context()).
		WithField("method", info.FullMethod).
		WithError(err)

	// handle errors happening when the client terminates the server-streaming RPC
	if errors.Is(err, context.Canceled) {
		log.Info("interrupted: context canceled")
		return nil
	}

	st := status.Convert(err)
	switch st.Code() {
	case codes.Canceled:
		log.Info("interrupted: gRPC operation canceled")
		return nil
	case codes.Unavailable:
		if st.Message() == "transport is closing" {
			log.Infof("interrupted: %s", st.Message())
			return nil
		}
	}

	// handle other errors
	log.Error("stream response failed")

	return err
}
