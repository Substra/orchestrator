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

package distributed

import (
	"github.com/owkin/orchestrator/lib/asset"
	"golang.org/x/net/context"
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

// RegisterDataSample will add a new DataSample to the state
func (a *DataSampleAdapter) RegisterDataSample(ctx context.Context, param *asset.NewDataSample) (*asset.NewDataSampleResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.datasample:RegisterDataSample"
	response := &asset.NewDataSampleResponse{}

	err = invocator.Invoke(method, param, response)

	return response, err
}

// UpdateDataSample will update a DataSample from the state
func (a *DataSampleAdapter) UpdateDataSample(ctx context.Context, param *asset.DataSampleUpdateParam) (*asset.DataSampleUpdateResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.datasample:UpdateDataSample"
	response := &asset.DataSampleUpdateResponse{}

	err = invocator.Invoke(method, param, response)

	return response, err
}

// QueryDataSamples returns all DataSamples
func (a *DataSampleAdapter) QueryDataSamples(ctx context.Context, param *asset.DataSamplesQueryParam) (*asset.DataSamplesQueryResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.datasample:QueryDataSamples"
	response := &asset.DataSamplesQueryResponse{}

	err = invocator.Invoke(method, param, response)

	return response, err
}
