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

package chaincode

import (
	"github.com/owkin/orchestrator/lib/assets"
	"golang.org/x/net/context"
)

// ObjectiveAdapter is a grpc server exposing the same node interface,
// but relies on a remote chaincode to actually manage the assets.
type ObjectiveAdapter struct {
	assets.UnimplementedObjectiveServiceServer
}

// NewObjectiveAdapter creates a Server
func NewObjectiveAdapter() *ObjectiveAdapter {
	return &ObjectiveAdapter{}
}

// RegisterObjective will add a new Objective to the network
func (a *ObjectiveAdapter) RegisterObjective(ctx context.Context, in *assets.NewObjective) (*assets.Objective, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.objective:RegisterObjective"

	response := &assets.Objective{}

	err = invocator.Invoke(method, in, response)

	return response, err
}

// QueryObjective returns an objective from its key
func (a *ObjectiveAdapter) QueryObjective(ctx context.Context, query *assets.ObjectiveQueryParam) (*assets.Objective, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.objective:QueryObjective"

	response := &assets.Objective{}

	err = invocator.Invoke(method, query, response)

	return response, err
}

// QueryObjectives returns all known objectives
func (a *ObjectiveAdapter) QueryObjectives(ctx context.Context, query *assets.ObjectivesQueryParam) (*assets.ObjectivesQueryResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.objective:QueryObjectives"

	response := &assets.ObjectivesQueryResponse{}

	err = invocator.Invoke(method, query, response)

	return response, err
}
