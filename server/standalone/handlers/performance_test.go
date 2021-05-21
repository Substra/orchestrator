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
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/standalone/concurrency"
	"github.com/stretchr/testify/assert"
)

func TestPerformanceServiceServer(t *testing.T) {
	server := NewPerformanceServer(new(concurrency.ImmediateRequestScheduler))
	assert.Implements(t, (*asset.PerformanceServiceServer)(nil), server)
}

func TestRegisterPerformance(t *testing.T) {
	ctx, p := getContext()
	ps := new(service.MockPerformanceService)

	server := NewPerformanceServer(new(concurrency.ImmediateRequestScheduler))

	newPerf := &asset.NewPerformance{ComputeTaskKey: "uuid", PerformanceValue: 3.14}

	p.On("GetPerformanceService").Return(ps)
	ps.On("RegisterPerformance", newPerf, "requester").Once().Return(&asset.Performance{ComputeTaskKey: "uuid"}, nil)

	_, err := server.RegisterPerformance(ctx, newPerf)
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ps.AssertExpectations(t)
}

func TestGetPerformance(t *testing.T) {
	ctx, p := getContext()
	ps := new(service.MockPerformanceService)

	server := NewPerformanceServer(new(concurrency.ImmediateRequestScheduler))

	perf := &asset.Performance{ComputeTaskKey: "uuid", PerformanceValue: 3.14}

	p.On("GetPerformanceService").Return(ps)
	ps.On("GetComputeTaskPerformance", "uuid").Once().Return(perf, nil)

	_, err := server.GetComputeTaskPerformance(ctx, &asset.GetComputeTaskPerformanceParam{ComputeTaskKey: "uuid"})
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ps.AssertExpectations(t)
}
