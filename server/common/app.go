package common

import (
	"time"

	"google.golang.org/grpc"
)

// Runnable is the opaque interface behind which standalone and distributed servers are handled
type Runnable interface {
	GetGrpcServer() *grpc.Server
	Stop()
}

// AppParameters are settings used by both distributed and standalone applications.
type AppParameters struct {
	GrpcOptions []grpc.ServerOption
	Config      *OrchestratorConfiguration
	RetryBudget time.Duration
}
