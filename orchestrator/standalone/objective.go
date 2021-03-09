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
	"log"

	"context"

	"github.com/owkin/orchestrator/lib/assets"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/orchestrator/common"
)

// ObjectiveServer is the gRPC facade to Objective manipulation
type ObjectiveServer struct {
	assets.UnimplementedObjectiveServiceServer
}

// NewObjectiveServer creates a grpc server
func NewObjectiveServer() *ObjectiveServer {
	return &ObjectiveServer{}
}

// RegisterObjective will persiste a new objective
func (s *ObjectiveServer) RegisterObjective(ctx context.Context, o *assets.NewObjective) (*assets.Objective, error) {
	log.Println(o)
	log.Printf("objective: %s, %s, %s", o.GetKey(), o.GetName(), o.GetTestDataset())

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetObjectiveService().RegisterObjective(o, mspid)
}

// QueryObjective fetches an objective by its key
func (s *ObjectiveServer) QueryObjective(ctx context.Context, params *assets.ObjectiveQueryParam) (*assets.Objective, error) {
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetObjectiveService().GetObjective(params.Key)
}

// QueryObjectives returns a paginated list of all known objectives
func (s *ObjectiveServer) QueryObjectives(ctx context.Context, params *assets.ObjectivesQueryParam) (*assets.ObjectivesQueryResponse, error) {
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	objectives, paginationToken, err := services.GetObjectiveService().GetObjectives(libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &assets.ObjectivesQueryResponse{
		Objectives:    objectives,
		NextPageToken: paginationToken,
	}, nil
}
