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

func TestAlgoAdapterImplementServer(t *testing.T) {
	adapter := NewAlgoAdapter()
	assert.Implementsf(t, (*asset.AlgoServiceServer)(nil), adapter, "AlgoAdapter should implements AlgoServiceServer")
}

func TestRegisterAlgo(t *testing.T) {
	adapter := NewAlgoAdapter()

	newObj := &asset.NewAlgo{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.algo:RegisterAlgo", newObj, &asset.Algo{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterAlgo(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestQueryAlgo(t *testing.T) {
	adapter := NewAlgoAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.AlgoQueryParam{Key: "uuid"}

	invocator.On("Call", "orchestrator.algo:QueryAlgo", param, &asset.Algo{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryAlgo(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryAlgos(t *testing.T) {
	adapter := NewAlgoAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.AlgosQueryParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", "orchestrator.algo:QueryAlgos", param, &asset.AlgosQueryResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryAlgos(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
