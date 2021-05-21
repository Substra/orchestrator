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

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"

	"github.com/owkin/orchestrator/server/standalone/concurrency"
	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// PerformanceServer is the gRPC facade to Performance manipulation
type PerformanceServer struct {
	asset.UnimplementedPerformanceServiceServer
	scheduler concurrency.RequestScheduler
}

// NewPerformanceServer creates a grpc server
func NewPerformanceServer(scheduler concurrency.RequestScheduler) *PerformanceServer {
	return &PerformanceServer{scheduler: scheduler}
}

func (s *PerformanceServer) RegisterPerformance(ctx context.Context, newPerf *asset.NewPerformance) (*asset.Performance, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetPerformanceService().RegisterPerformance(newPerf, mspid)
}

func (s *PerformanceServer) GetComputeTaskPerformance(ctx context.Context, param *asset.GetComputeTaskPerformanceParam) (*asset.Performance, error) {
	execToken := <-s.scheduler.AcquireExecutionToken()
	defer execToken.Release()

	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetPerformanceService().GetComputeTaskPerformance(param.ComputeTaskKey)
}
