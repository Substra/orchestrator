package common

import (
	"context"
	"strings"

	"github.com/go-playground/log/v7"
	"google.golang.org/grpc"
)

// LogRequest log every gRPC response
func LogRequest(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	logger := log.WithField("method", info.FullMethod)

	resp, err := handler(ctx, req)

	if err != nil {
		logger.WithError(err).Error("Error response")
	} else {
		logger.WithField("response", resp).Debug("Success reponse")
	}

	return resp, err
}
