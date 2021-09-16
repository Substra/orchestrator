package common

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/server/common/logger"
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

	log := logger.Get(ctx).WithField("method", info.FullMethod)

	resp, err := handler(ctx, req)

	if err == nil {
		// Error is already logged by the error interceptor
		log.WithField("response", resp).Debug("Success response")
	}

	return resp, err
}
