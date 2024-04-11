package handlers

import (
	"context"

	"github.com/substra/orchestrator/lib/asset"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/substra/orchestrator/server/standalone/interceptors"
)

// ProfilingServer is the gRPC facade to Profiling manipulation
type ProfilingServer struct {
	asset.UnimplementedProfilingServiceServer
}

// NewProfilingServer creates a grpc server
func NewProfilingServer() *ProfilingServer {
	return &ProfilingServer{}
}

func (s *ProfilingServer) RegisterProfilingStep(ctx context.Context, ps *asset.ProfilingStep) (*emptypb.Empty, error) {
	_, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return nil, services.GetProfilingService().RegisterProfilingStep(ps)
}
