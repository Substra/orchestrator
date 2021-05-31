// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"

	"github.com/go-playground/log/v7"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"

	"github.com/owkin/orchestrator/server/standalone/concurrency"
	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// DataSampleServer is the gRPC facade to DataSample manipulation
type DataSampleServer struct {
	asset.UnimplementedDataSampleServiceServer
	scheduler concurrency.RequestScheduler
}

// NewDataSampleServer creates a gRPC server
func NewDataSampleServer(scheduler concurrency.RequestScheduler) *DataSampleServer {
	return &DataSampleServer{scheduler: scheduler}
}

// RegisterDataSample will persist new DataSamples
func (s *DataSampleServer) RegisterDataSamples(ctx context.Context, input *asset.RegisterDataSamplesParam) (*asset.RegisterDataSamplesResponse, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

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
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

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
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

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
