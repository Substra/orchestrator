package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
)

// InfoServer is the gRPC server exposing info actions
type InfoServer struct {
	asset.UnimplementedInfoServiceServer
}

// NewInfoServer creates a Server
func NewInfoServer() *InfoServer {
	return &InfoServer{}
}

func (s *InfoServer) QueryVersion(ctx context.Context, in *asset.QueryVersionParam) (*asset.QueryVersionResponse, error) {
	return &asset.QueryVersionResponse{
		Orchestrator: common.Version,
	}, nil
}
