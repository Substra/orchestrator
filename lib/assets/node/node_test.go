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

package node

import (
	"testing"

	"github.com/stretchr/testify/mock"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
)

func TestRegistration(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	node := Node{
		Id: "uuid1",
	}

	mockDB.On("PutState", resource, "uuid1", mock.Anything).Return(nil).Once()

	service := NewService(mockDB)

	service.RegisterNode(&node)
}
