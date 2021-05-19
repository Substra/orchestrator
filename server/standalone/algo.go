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
	scheduler RequestScheduler
}

// NewAlgoServer creates a grpc server
func NewAlgoServer(scheduler RequestScheduler) *AlgoServer {
	return &AlgoServer{scheduler: scheduler}
}

// RegisterAlgo will persiste a new algo
func (s *AlgoServer) RegisterAlgo(ctx context.Context, a *asset.NewAlgo) (*asset.Algo, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

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

// GetAlgo fetches an algo by its key
func (s *AlgoServer) GetAlgo(ctx context.Context, params *asset.GetAlgoParam) (*asset.Algo, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetAlgoService().GetAlgo(params.Key)
}

// QueryAlgos returns a paginated list of all known algos
func (s *AlgoServer) QueryAlgos(ctx context.Context, params *asset.QueryAlgosParam) (*asset.QueryAlgosResponse, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	algos, paginationToken, err := services.GetAlgoService().QueryAlgos(params.Category, libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryAlgosResponse{
		Algos:         algos,
		NextPageToken: paginationToken,
	}, nil
}
