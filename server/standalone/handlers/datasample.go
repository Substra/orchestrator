package handlers

import (
	"context"

	"github.com/go-playground/log/v7"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// DataSampleServer is the gRPC facade to DataSample manipulation
type DataSampleServer struct {
	asset.UnimplementedDataSampleServiceServer
}

// NewDataSampleServer creates a gRPC server
func NewDataSampleServer() *DataSampleServer {
	return &DataSampleServer{}
}

// RegisterDataSamples will persist new DataSamples
func (s *DataSampleServer) RegisterDataSamples(ctx context.Context, input *asset.RegisterDataSamplesParam) (*asset.RegisterDataSamplesResponse, error) {
	log.WithField("nbSamples", len(input.Samples)).Debug("Register DataSamples")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetDataSampleService().RegisterDataSamples(input.Samples, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.RegisterDataSamplesResponse{}, nil
}

// UpdateDataSamples will update the datamanagers existing DataSamples
func (s *DataSampleServer) UpdateDataSamples(ctx context.Context, datasample *asset.UpdateDataSamplesParam) (*asset.UpdateDataSamplesResponse, error) {
	log.WithField("datasample", datasample).Debug("Update DataSample")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetDataSampleService().UpdateDataSamples(datasample, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.UpdateDataSamplesResponse{}, nil
}

// QueryDataSamples returns a paginated list of all known datasamples
func (s *DataSampleServer) QueryDataSamples(ctx context.Context, params *asset.QueryDataSamplesParam) (*asset.QueryDataSamplesResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	datasamples, paginationToken, err := services.GetDataSampleService().QueryDataSamples(libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryDataSamplesResponse{
		DataSamples:   datasamples,
		NextPageToken: paginationToken,
	}, nil
}
