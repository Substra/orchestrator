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

// ObjectiveAdapter is a grpc server exposing the same node interface,
// but relies on a remote chaincode to actually manage the asset.
type ObjectiveAdapter struct {
	asset.UnimplementedObjectiveServiceServer
}

// NewObjectiveAdapter creates a Server
func NewObjectiveAdapter() *ObjectiveAdapter {
	return &ObjectiveAdapter{}
}

// RegisterObjective will add a new Objective to the network
func (a *ObjectiveAdapter) RegisterObjective(ctx context.Context, in *asset.NewObjective) (*asset.Objective, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.objective:RegisterObjective"

	response := &asset.Objective{}

	err = invocator.Call(method, in, response)

	return response, err
}

// GetObjective returns an objective from its key
func (a *ObjectiveAdapter) GetObjective(ctx context.Context, query *asset.GetObjectiveParam) (*asset.Objective, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.objective:GetObjective"

	response := &asset.Objective{}

	err = invocator.Call(method, query, response)

	return response, err
}

// QueryObjectives returns all known objectives
func (a *ObjectiveAdapter) QueryObjectives(ctx context.Context, query *asset.QueryObjectivesParam) (*asset.QueryObjectivesResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.objective:QueryObjectives"

	response := &asset.QueryObjectivesResponse{}

	err = invocator.Call(method, query, response)

	return response, err
}
