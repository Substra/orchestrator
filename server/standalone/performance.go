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

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
)

// PerformanceServer is the gRPC facade to Performance manipulation
type PerformanceServer struct {
	asset.UnimplementedPerformanceServiceServer
}

// NewPerformanceServer creates a grpc server
func NewPerformanceServer() *PerformanceServer {
	return &PerformanceServer{}
}

func (s *PerformanceServer) RegisterPerformance(ctx context.Context, newPerf *asset.NewPerformance) (*asset.Performance, error) {
	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetPerformanceService().RegisterPerformance(newPerf, mspid)
}

func (s *PerformanceServer) GetComputeTaskPerformance(ctx context.Context, param *asset.GetComputeTaskPerformanceParam) (*asset.Performance, error) {
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetPerformanceService().GetComputeTaskPerformance(param.ComputeTaskKey)
}
