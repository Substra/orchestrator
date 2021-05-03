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

func TestObjectiveAdapterImplementServer(t *testing.T) {
	adapter := NewObjectiveAdapter()
	assert.Implementsf(t, (*asset.ObjectiveServiceServer)(nil), adapter, "ObjectiveAdapter should implements ObjectiveServiceServer")
}

func TestRegisterObjective(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newObj := &asset.NewObjective{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.objective:RegisterObjective", newObj, &asset.Objective{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterObjective(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestQueryObjective(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.ObjectiveQueryParam{Key: "uuid"}

	invocator.On("Call", "orchestrator.objective:QueryObjective", param, &asset.Objective{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryObjective(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryObjectives(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.ObjectivesQueryParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", "orchestrator.objective:QueryObjectives", param, &asset.ObjectivesQueryResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryObjectives(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
