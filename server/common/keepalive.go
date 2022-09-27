package common

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultMinTime             = "30s"
	defaultPermitWithoutStream = "false"
)

// GetKeepAliveOptions will return server option with gRPC keepalive setup.
// This may panic on missing or invalid configuration env var.
func GetKeepAliveOptions() grpc.ServerOption {
	minTime := MustParseDuration(GetEnvOrFallback("GRPC_KEEPALIVE_POLICY_MIN_TIME", defaultMinTime))
	permitWithoutStream := MustParseBool(GetEnvOrFallback("GRPC_KEEPALIVE_POLICY_PERMIT_WITHOUT_STREAM", defaultPermitWithoutStream))

	return grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             minTime,
		PermitWithoutStream: permitWithoutStream,
	})
}
