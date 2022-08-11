package interceptors

import (
	"context"
	"strings"

	"errors"

	"github.com/rs/zerolog/log"
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

	logger := log.Ctx(ctx).With().Str("method", info.FullMethod).Logger()

	resp, err := handler(ctx, req)

	if err == nil {
		// Error is already logged by the error interceptor
		logger.Debug().Interface("response", resp).Msg("Success response")
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

	logger := log.Ctx(stream.Context()).
		With().
		Str("method", info.FullMethod).
		Err(err).
		Logger()

	// handle errors happening when the client terminates the server-streaming RPC
	if errors.Is(err, context.Canceled) {
		logger.Info().Msg("interrupted: context canceled")
		return nil
	}

	st := status.Convert(err)
	switch st.Code() {
	case codes.Canceled:
		logger.Info().Msg("interrupted: gRPC operation canceled")
		return nil
	case codes.Unavailable:
		if st.Message() == "transport is closing" {
			logger.Info().Msgf("interrupted: %s", st.Message())
			return nil
		}
	}

	// handle other errors
	logger.Error().Msg("stream response failed")

	return err
}
