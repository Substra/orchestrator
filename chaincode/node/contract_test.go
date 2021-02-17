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

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/orchestration"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *testHelper.MockedContext) *orchestration.MockNodeService {
	mockService := new(orchestration.MockNodeService)

	provider := new(orchestration.MockServiceProvider)
	provider.On("GetNodeService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	o := &assets.Node{Id: org}
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterNode", org).Return(o, nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	node, err := contract.RegisterNode(ctx)
	assert.Nil(t, err, "node registration should not fail")

	assert.Equal(t, node, o)
}

func TestQueryNodes(t *testing.T) {
	contract := &SmartContract{}

	nodes := []*assets.Node{
		{Id: "org1"},
		{Id: "org2"},
	}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("GetNodes").Return(nodes, nil).Once()

	resp, err := contract.QueryNodes(ctx)
	assert.Nil(t, err, "querying nodes should not fail")
	assert.Len(t, resp, len(nodes), "query should return all nodes")
}
