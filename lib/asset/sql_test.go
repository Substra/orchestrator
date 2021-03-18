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
