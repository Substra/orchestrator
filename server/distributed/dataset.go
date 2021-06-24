package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// DatasetAdapter is a grpc server exposing the same dataset interface,
// but relies on a remote chaincode to actually manage the asset.
type DatasetAdapter struct {
	asset.UnimplementedDatasetServiceServer
}

// NewDatasetAdapter creates a Server
func NewDatasetAdapter() *DatasetAdapter {
	return &DatasetAdapter{}
}

// GetDataset fetches a dataset by its key
func (s *DatasetAdapter) GetDataset(ctx context.Context, params *asset.GetDatasetParam) (*asset.Dataset, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.dataset:GetDataset"
	response := &asset.Dataset{}

	err = invocator.Call(method, params, response)

	return response, err
}
