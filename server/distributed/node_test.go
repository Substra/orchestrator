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
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

// TestNodeAdapterImplementServer makes sure the chaincode-baked orchestration exposes the same server than standalone mode.
func TestNodeAdapterImplementServer(t *testing.T) {
	adapter := NewNodeAdapter()
	assert.Implementsf(t, (*asset.NodeServiceServer)(nil), adapter, "NodeAdapter should implements NodeServiceServer")
}
