package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// DatasetServer is the gRPC facade to Dataset manipulation
type DatasetServer struct {
	asset.UnimplementedDatasetServiceServer
}

// NewDatasetServer creates a gRPC server
func NewDatasetServer() *DatasetServer {
	return &DatasetServer{}
}

// GetDataset fetches a dataset by its key
func (s *DatasetServer) GetDataset(ctx context.Context, params *asset.GetDatasetParam) (*asset.Dataset, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetDatasetService().GetDataset(params.GetKey())
}
