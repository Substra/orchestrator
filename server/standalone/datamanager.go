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

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"
)

// DataManagerServer is the gRPC facade to DataManager manipulation
type DataManagerServer struct {
	asset.UnimplementedDataManagerServiceServer
	scheduler RequestScheduler
}

// NewDataManagerServer creates a gRPC server
func NewDataManagerServer(scheduler RequestScheduler) *DataManagerServer {
	return &DataManagerServer{scheduler: scheduler}
}

// RegisterDataManager will persist new DataManagers
func (s *DataManagerServer) RegisterDataManager(ctx context.Context, d *asset.NewDataManager) (*asset.NewDataManagerResponse, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	log.WithField("datamanager", d).Debug("Register DataManager")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetDataManagerService().RegisterDataManager(d, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.NewDataManagerResponse{}, nil
}

// UpdateDataManager will update the objective of an existing DataManager
func (s *DataManagerServer) UpdateDataManager(ctx context.Context, d *asset.DataManagerUpdateParam) (*asset.DataManagerUpdateResponse, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	log.WithField("datamanager", d).Debug("Update UpdateDataManager")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetDataManagerService().UpdateDataManager(d, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.DataManagerUpdateResponse{}, nil
}

// GetDataManager fetches a datamanager by its key
func (s *DataManagerServer) GetDataManager(ctx context.Context, params *asset.GetDataManagerParam) (*asset.DataManager, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetDataManagerService().GetDataManager(params.GetKey())
}

// QueryDataManagers returns a paginated list of all known datamanagers
func (s *DataManagerServer) QueryDataManagers(ctx context.Context, params *asset.QueryDataManagersParam) (*asset.QueryDataManagersResponse, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	datamanagers, paginationToken, err := services.GetDataManagerService().QueryDataManagers(libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryDataManagersResponse{
		DataManagers:  datamanagers,
		NextPageToken: paginationToken,
	}, nil
}
