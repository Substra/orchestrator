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

	"github.com/owkin/orchestrator/lib/assets"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegistration(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	service := NewService(mockDB)

	description := &assets.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	metrics := &assets.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	perms := &assets.Permissions{Process: &assets.Permission{Public: true}}

	objective := Objective{
		Key:         "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:        "Test objective",
		MetricsName: "test perf",
		Metrics:     metrics,
		Description: description,
		Permissions: perms,
	}

	mockDB.On("PutState", resource, "08680966-97ae-4573-8b2d-6c4db2b3c532", mock.Anything).Return(nil).Once()

	err := service.RegisterObjective(&objective)
	assert.NoError(t, err, "Registration of valid objective should not fail")

	mockDB.AssertExpectations(t)
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
