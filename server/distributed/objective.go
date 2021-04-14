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
	method := "org.substra.objective:RegisterObjective"

	response := &asset.Objective{}

	err = invocator.Invoke(method, in, response)

	return response, err
}

// QueryObjective returns an objective from its key
func (a *ObjectiveAdapter) QueryObjective(ctx context.Context, query *asset.ObjectiveQueryParam) (*asset.Objective, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.objective:QueryObjective"

	response := &asset.Objective{}

	err = invocator.Evaluate(method, query, response)

	return response, err
}

// QueryObjectives returns all known objectives
func (a *ObjectiveAdapter) QueryObjectives(ctx context.Context, query *asset.ObjectivesQueryParam) (*asset.ObjectivesQueryResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.objective:QueryObjectives"

	response := &asset.ObjectivesQueryResponse{}

	err = invocator.Evaluate(method, query, response)

	return response, err
}
