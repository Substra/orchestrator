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
	"github.com/stretchr/testify/mock"
)

func TestRegisterDatamanager(t *testing.T) {
	adapter := NewDataManagerAdapter()

	newObj := &asset.NewDataManager{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.datamanager:RegisterDataManager", newObj, &asset.DataManager{}).
		Once().
		Run(func(args mock.Arguments) {
			dm := args.Get(2).(*asset.DataManager)
			dm.Key = "uuid"
			dm.Owner = "test"
		}).
		Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	dm, err := adapter.RegisterDataManager(ctx, newObj)
	assert.NoError(t, err, "Registration should pass")

	assert.Equal(t, "uuid", dm.Key)
	assert.Equal(t, "test", dm.Owner)
}
