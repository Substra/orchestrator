// Copyright 2020 Owkin Inc.
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

package orchestration

import (
	"testing"

	"github.com/owkin/orchestrator/lib/assets"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterNode(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	node := assets.Node{
		Id: "uuid1",
	}

	mockDB.On("PutState", nodeResource, "uuid1", mock.Anything).Return(nil).Once()

	service := NewNodeService(mockDB)

	err := service.RegisterNode(&node)
	assert.NoError(t, err, "Node registration should not fail")
}
