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
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestPerformanceAdapterImplementServer(t *testing.T) {
	adapter := NewPerformanceAdapter()
	assert.Implements(t, (*asset.PerformanceServiceServer)(nil), adapter)
}

func TestRegisterPerformance(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.NewPerformance{}

	invocator.On("Call", "orchestrator.performance:RegisterPerformance", param, &asset.Performance{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterPerformance(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestGetPerformance(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.GetComputeTaskPerformanceParam{}

	invocator.On("Call", "orchestrator.performance:GetComputeTaskPerformance", param, &asset.Performance{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.GetComputeTaskPerformance(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
