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

package chaincode

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/owkin/orchestrator/lib/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectiveAdapterImplementServer(t *testing.T) {
	adapter := NewObjectiveAdapter()
	assert.Implementsf(t, (*assets.ObjectiveServiceServer)(nil), adapter, "ObjectiveAdapter should implements ObjectiveServiceServer")
}

func TestRegisterObjective(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newObj := &assets.NewObjective{
		Key: "uuid",
	}
	oBytes, err := json.Marshal(newObj)
	require.NoError(t, err)

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	params := []string{string(oBytes)}

	invocator.On("Invoke", "org.substra.objective:RegisterObjective", params, &assets.Objective{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err = adapter.RegisterObjective(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}
