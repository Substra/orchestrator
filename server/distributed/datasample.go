package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// DataSampleAdapter is a grpc server exposing the same datasample interface,
// but relies on a remote chaincode to actually manage the asset.
type DataSampleAdapter struct {
	asset.UnimplementedDataSampleServiceServer
}

// NewDataSampleAdapter creates a Server
func NewDataSampleAdapter() *DataSampleAdapter {
	return &DataSampleAdapter{}
}

// RegisterDataSamples will add a new DataSample to the state
func (a *DataSampleAdapter) RegisterDataSamples(ctx context.Context, param *asset.RegisterDataSamplesParam) (*asset.RegisterDataSamplesResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datasample:RegisterDataSamples"

	err = invocator.Call(ctx, method, param, nil)

	return &asset.RegisterDataSamplesResponse{}, err
}

// UpdateDataSamples will update a DataSample from the state
func (a *DataSampleAdapter) UpdateDataSamples(ctx context.Context, param *asset.UpdateDataSamplesParam) (*asset.UpdateDataSamplesResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datasample:UpdateDataSamples"

	err = invocator.Call(ctx, method, param, nil)

	return &asset.UpdateDataSamplesResponse{}, err
}

// QueryDataSamples returns all DataSamples
func (a *DataSampleAdapter) QueryDataSamples(ctx context.Context, param *asset.QueryDataSamplesParam) (*asset.QueryDataSamplesResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datasample:QueryDataSamples"
	response := &asset.QueryDataSamplesResponse{}

	err = invocator.Call(ctx, method, param, response)

	return response, err
}
