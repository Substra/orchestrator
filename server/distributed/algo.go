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

// AlgoAdapter is a grpc server exposing the same algo interface than standalone,
// but relies on a remote chaincode to actually manage the asset.
type AlgoAdapter struct {
	asset.UnimplementedAlgoServiceServer
}

// NewAlgoAdapter creates a Server
func NewAlgoAdapter() *AlgoAdapter {
	return &AlgoAdapter{}
}

// RegisterAlgo will add a new Algo to the network
func (a *AlgoAdapter) RegisterAlgo(ctx context.Context, in *asset.NewAlgo) (*asset.Algo, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.algo:RegisterAlgo"

	response := &asset.Algo{}

	err = invocator.Call(method, in, response)

	return response, err
}

// GetAlgo returns an algo from its key
func (a *AlgoAdapter) GetAlgo(ctx context.Context, query *asset.GetAlgoParam) (*asset.Algo, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.algo:GetAlgo"

	response := &asset.Algo{}

	err = invocator.Call(method, query, response)

	return response, err
}

// QueryAlgos returns all known algos
func (a *AlgoAdapter) QueryAlgos(ctx context.Context, query *asset.QueryAlgosParam) (*asset.QueryAlgosResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.algo:QueryAlgos"

	response := &asset.QueryAlgosResponse{}

	err = invocator.Call(method, query, response)

	return response, err
}
