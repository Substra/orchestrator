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

package standalone

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
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
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetDatasetService().GetDataset(params.GetKey())
}
