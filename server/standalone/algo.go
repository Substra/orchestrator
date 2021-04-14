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
	"github.com/go-playground/log/v7"

	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"
)

// AlgoServer is the gRPC facade to Algo manipulation
type AlgoServer struct {
	asset.UnimplementedAlgoServiceServer
}

// NewAlgoServer creates a grpc server
func NewAlgoServer() *AlgoServer {
	return &AlgoServer{}
}

// RegisterAlgo will persiste a new algo
func (s *AlgoServer) RegisterAlgo(ctx context.Context, a *asset.NewAlgo) (*asset.Algo, error) {
	log.WithField("algo", a).Debug("Register Algo")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetAlgoService().RegisterAlgo(a, mspid)
}

// QueryAlgo fetches an algo by its key
func (s *AlgoServer) QueryAlgo(ctx context.Context, params *asset.AlgoQueryParam) (*asset.Algo, error) {
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetAlgoService().GetAlgo(params.Key)
}

// QueryAlgos returns a paginated list of all known algos
func (s *AlgoServer) QueryAlgos(ctx context.Context, params *asset.AlgosQueryParam) (*asset.AlgosQueryResponse, error) {
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	algos, paginationToken, err := services.GetAlgoService().GetAlgos(params.Category, libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.AlgosQueryResponse{
		Algos:         algos,
		NextPageToken: paginationToken,
	}, nil
}
