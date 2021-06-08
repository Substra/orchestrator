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
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// DataManagerAdapter is a grpc server exposing the same dataManager interface,
// but relies on a remote chaincode to actually manage the asset.
type DataManagerAdapter struct {
	asset.UnimplementedDataManagerServiceServer
}

// NewDataManagerAdapter creates a Server
func NewDataManagerAdapter() *DataManagerAdapter {
	return &DataManagerAdapter{}
}

// RegisterDataManager will persist new DataManagers
func (s *DataManagerAdapter) RegisterDataManager(ctx context.Context, d *asset.NewDataManager) (*asset.DataManager, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:RegisterDataManager"

	response := &asset.DataManager{}

	err = invocator.Call(method, d, response)

	return response, err
}

// UpdateDataManager will update the objective of an existing DataManager
func (s *DataManagerAdapter) UpdateDataManager(ctx context.Context, d *asset.DataManagerUpdateParam) (*asset.DataManagerUpdateResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:UpdateDataManager"

	err = invocator.Call(method, d, nil)

	return &asset.DataManagerUpdateResponse{}, err
}

// GetDataManager fetches a datamanager by its key
func (s *DataManagerAdapter) GetDataManager(ctx context.Context, params *asset.GetDataManagerParam) (*asset.DataManager, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:GetDataManager"
	response := &asset.DataManager{}

	err = invocator.Call(method, params, response)

	return response, err
}

// QueryDataManagers returns a paginated list of all known datamanagers
func (s *DataManagerAdapter) QueryDataManagers(ctx context.Context, params *asset.QueryDataManagersParam) (*asset.QueryDataManagersResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:QueryDataManagers"
	response := &asset.QueryDataManagersResponse{}

	err = invocator.Call(method, params, response)

	return response, err
}
