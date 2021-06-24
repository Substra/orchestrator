package common

import "google.golang.org/grpc"

// Runnable is the opaque interface behind which standalone and distributed servers are handled
type Runnable interface {
	GetGrpcServer() *grpc.Server
	Stop()
}
