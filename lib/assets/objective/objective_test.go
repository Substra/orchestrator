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

package objective

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	persistenceHelper "github.com/substrafoundation/substra-orchestrator/lib/persistence/testing"
)

func TestRegistration(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	service := NewService(mockDB)

	objective := Objective{
		Key: "objKey",
	}

	mockDB.On("PutState", resource, "objKey", mock.Anything).Return(nil).Once()

	service.RegisterObjective(&objective)
}

func TestQuery(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	service := NewService(mockDB)

	objective := Objective{
		Key:  "objKey",
		Name: "Test",
	}

	objBytes, err := json.Marshal(&objective)
	require.Nil(t, err)

	mockDB.On("GetState", resource, "objKey").Return(objBytes, nil).Once()

	o, err := service.GetObjective("objKey")
	require.Nil(t, err)
	assert.Equal(t, o.Name, objective.Name)
}

func TestGetObjectives(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	service := NewService(mockDB)

	obj1 := Objective{
		Key:  "obj1",
		Name: "Test 1",
	}
	obj2 := Objective{
		Key:  "obj2",
		Name: "Test 2",
	}

	bytes1, err := json.Marshal(&obj1)
	require.Nil(t, err)
	bytes2, err := json.Marshal(&obj2)
	require.Nil(t, err)

	mockDB.On("GetAll", resource).Return([][]byte{bytes1, bytes2}, nil).Once()

	r, err := service.GetObjectives()
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, obj1.Key)
}
