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

package orchestration

import (
	"testing"

	"github.com/owkin/orchestrator/lib/assets"
	"github.com/stretchr/testify/assert"
)

func TestNewPermission(t *testing.T) {
	n := assets.NewPermissions{Public: false}

	p := newPermission(&n, "org")

	assert.Contains(t, p.AuthorizedIds, "org", "owner should be added to authorized IDs")
}

func TestValidateAuthorizedIDs(t *testing.T) {
	mockNodeService := new(mockNodeService)
	provider := new(mockServiceProvider)
	provider.On("GetNodeService").Return(mockNodeService)
	service := NewPermissionService(provider)

	nodes := []*assets.Node{
		{Id: "org1"},
		{Id: "org2"},
	}
	mockNodeService.On("GetNodes").Return(nodes, nil)

	assert.Error(t, service.validateAuthorizedIDs([]string{"orgA"}), "orgA is not a valid node")
	assert.NoError(t, service.validateAuthorizedIDs([]string{"org1"}), "org1 is a valid node")
}
