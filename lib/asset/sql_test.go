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

package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectiveValue(t *testing.T) {
	objective := &Objective{
		Name:  "test",
		Owner: "testOwner",
	}

	value, err := objective.Value()
	assert.NoError(t, err, "objective serialization should not fail")

	scanned := new(Objective)
	err = scanned.Scan(value)
	assert.NoError(t, err, "objective scan should not fail")

	assert.Equal(t, objective, scanned)
}

func TestDataSampleValue(t *testing.T) {
	datasample := &DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "testOwner",
		TestOnly:        false,
	}

	value, err := datasample.Value()
	assert.NoError(t, err, "datasample serialization should not fail")

	scanned := new(DataSample)
	err = scanned.Scan(value)
	assert.NoError(t, err, "datasample scan should not fail")

	assert.Equal(t, datasample, scanned)
}

func TestAlgoValue(t *testing.T) {
	algo := &Algo{
		Name:  "test",
		Owner: "testOwner",
	}

	value, err := algo.Value()
	assert.NoError(t, err, "algo serialization should not fail")

	scanned := new(Algo)
	err = scanned.Scan(value)
	assert.NoError(t, err, "algo scan should not fail")

	assert.Equal(t, algo, scanned)
}

func TestDataManagerValue(t *testing.T) {
	datamanager := &DataManager{
		Name:  "test",
		Owner: "testOwner",
	}

	value, err := datamanager.Value()
	assert.NoError(t, err, "datamanager serialization should not fail")

	scanned := new(DataManager)
	err = scanned.Scan(value)
	assert.NoError(t, err, "datamanager scan should not fail")

	assert.Equal(t, datamanager, scanned)
}

func TestModelValue(t *testing.T) {
	model := &Model{
		Key:            "08bcb3b9-015c-4b6a-a9b5-033b3b324a7c",
		Category:       ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey: "08bcb3b9-015c-4b6a-a9b5-033b3b324a7d",
		Address: &Addressable{
			Checksum:       "c15918f80d920769904e92bae59cf4b926b362201ad686c2834403a08a19de16",
			StorageAddress: "http://somewhere.online",
		},
	}

	value, err := model.Value()
	assert.NoError(t, err, "model serialization should not fail")

	scanned := new(Model)
	err = scanned.Scan(value)
	assert.NoError(t, err, "model scan should not fail")

	assert.Equal(t, model, scanned)
}

func TestComputePlanValue(t *testing.T) {
	computeplan := &ComputePlan{
		Key: "08bcb3b9-015c-4b6a-a9b5-033b3b324a7c",
		Metadata: map[string]string{
			"test": "true",
		},
	}

	value, err := computeplan.Value()
	assert.NoError(t, err, "compute plan serialization should not fail")

	scanned := new(ComputePlan)
	err = scanned.Scan(value)
	assert.NoError(t, err, "compute plan scan should not fail")

	assert.Equal(t, computeplan, scanned)
}
