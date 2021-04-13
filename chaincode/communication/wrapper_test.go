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

package communication

import (
	"encoding/json"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestWrapUnwrap(t *testing.T) {
	msg := &asset.NewAlgo{
		Key:      "uuid",
		Category: asset.AlgoCategory_ALGO_SIMPLE,
	}

	wrapped, err := Wrap(msg)
	assert.NoError(t, err)

	out := new(asset.NewAlgo)
	err = wrapped.Unwrap(out)
	assert.NoError(t, err)
	assert.Equal(t, msg, out)

	serialized, err := json.Marshal(wrapped)
	assert.NoError(t, err)

	out = new(asset.NewAlgo)
	err = Unwrap(serialized, out)
	assert.NoError(t, err)
	assert.Equal(t, msg, out)
}
