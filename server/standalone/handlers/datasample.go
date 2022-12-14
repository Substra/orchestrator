package handlers

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	libCommon "github.com/substra/orchestrator/lib/common"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/interceptors"
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
	log.Ctx(ctx).Debug().Int("nbSamples", len(input.Samples)).Msg("Register DataSamples")

	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	datasamples, err := services.GetDataSampleService().RegisterDataSamples(input.Samples, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.RegisterDataSamplesResponse{DataSamples: datasamples}, nil

}

// UpdateDataSamples will update the datamanagers existing DataSamples
func (s *DataSampleServer) UpdateDataSamples(ctx context.Context, datasample *asset.UpdateDataSamplesParam) (*asset.UpdateDataSamplesResponse, error) {
	log.Ctx(ctx).Debug().Interface("datasample", datasample).Msg("Update DataSample")

	mspid, err := commonInterceptors.ExtractMSPID(ctx)
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

	datasamples, paginationToken, err := services.GetDataSampleService().QueryDataSamples(libCommon.NewPagination(params.PageToken, params.PageSize), params.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryDataSamplesResponse{
		DataSamples:   datasamples,
		NextPageToken: paginationToken,
	}, nil
}

// GetDataSample fetches a datasample by its key
func (s *DataSampleServer) GetDataSample(ctx context.Context, params *asset.GetDataSampleParam) (*asset.DataSample, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetDataSampleService().GetDataSample(params.Key)
}
