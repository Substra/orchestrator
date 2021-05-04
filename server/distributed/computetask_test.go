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

func TestComputeTaskAdapterImplementServer(t *testing.T) {
	adapter := NewComputeTaskAdapter()
	assert.Implements(t, (*asset.ComputeTaskServiceServer)(nil), adapter)
}

func TestRegisterTask(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.NewComputeTask{}

	invocator.On("Call", "orchestrator.computetask:RegisterTask", param, &asset.ComputeTask{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterTask(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryTasks(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.TasksQueryParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", "orchestrator.computetask:QueryTasks", param, &asset.TasksQueryResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryTasks(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
