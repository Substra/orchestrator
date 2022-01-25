package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
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

	response := &asset.RegisterDataSamplesResponse{}

	err = invocator.Call(ctx, method, param, response)

	if err != nil && isFabricTimeoutRetry(ctx) && len(param.Samples) == 1 && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout and the registration is limited to a single sample.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		datasample := &asset.DataSample{}
		err = invocator.Call(ctx, "orchestrator.datasample:GetDataSample", &asset.GetDataSampleParam{Key: param.Samples[0].Key}, datasample)

		return &asset.RegisterDataSamplesResponse{DataSamples: []*asset.DataSample{datasample}}, err
	}

	if err != nil {
		return nil, err
	}

	return response, nil
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

// GetDataSample returns a datasample from its key
func (a *DataSampleAdapter) GetDataSample(ctx context.Context, param *asset.GetDataSampleParam) (*asset.DataSample, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datasample:GetDataSample"

	response := &asset.DataSample{}

	err = invocator.Call(ctx, method, param, response)

	return response, err
}
