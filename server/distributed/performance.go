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

// PerformanceAdapter is a grpc server exposing the same Performance interface than standalone mode,
// but relies on a remote chaincode to actually manage the asset.
type PerformanceAdapter struct {
	asset.UnimplementedPerformanceServiceServer
}

// NewPerformanceAdapter creates a Server
func NewPerformanceAdapter() *PerformanceAdapter {
	return &PerformanceAdapter{}
}

func (a *PerformanceAdapter) RegisterPerformance(ctx context.Context, newPerf *asset.NewPerformance) (*asset.Performance, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.performance:RegisterPerformance"

	perf := &asset.Performance{}

	err = invocator.Call(method, newPerf, perf)

	return perf, err
}

func (a *PerformanceAdapter) GetComputeTaskPerformance(ctx context.Context, param *asset.GetComputeTaskPerformanceParam) (*asset.Performance, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.performance:GetComputeTaskPerformance"

	perf := &asset.Performance{}

	err = invocator.Call(method, param, perf)

	return perf, err
}
