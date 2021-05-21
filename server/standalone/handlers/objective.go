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

// ObjectiveServer is the gRPC facade to Objective manipulation
type ObjectiveServer struct {
	asset.UnimplementedObjectiveServiceServer
	scheduler concurrency.RequestScheduler
}

// NewObjectiveServer creates a grpc server
func NewObjectiveServer(scheduler concurrency.RequestScheduler) *ObjectiveServer {
	return &ObjectiveServer{scheduler: scheduler}
}

// RegisterObjective will persiste a new objective
func (s *ObjectiveServer) RegisterObjective(ctx context.Context, o *asset.NewObjective) (*asset.Objective, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	log.WithField("objective", o).Debug("register objective")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetObjectiveService().RegisterObjective(o, mspid)
}

// GetObjective fetches an objective by its key
func (s *ObjectiveServer) GetObjective(ctx context.Context, params *asset.GetObjectiveParam) (*asset.Objective, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetObjectiveService().GetObjective(params.Key)
}

// QueryObjectives returns a paginated list of all known objectives
func (s *ObjectiveServer) QueryObjectives(ctx context.Context, params *asset.QueryObjectivesParam) (*asset.QueryObjectivesResponse, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	objectives, paginationToken, err := services.GetObjectiveService().QueryObjectives(libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryObjectivesResponse{
		Objectives:    objectives,
		NextPageToken: paginationToken,
	}, nil
}
