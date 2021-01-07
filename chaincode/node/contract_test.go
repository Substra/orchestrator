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

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/msp"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockedService struct {
	mock.Mock
}

func (m *MockedService) RegisterNode(n *assets.Node) error {
	args := m.Called(n)
	return args.Error(0)
}

func (m *MockedService) GetNodes() ([]*assets.Node, error) {
	args := m.Called()
	return args.Get(0).([]*assets.Node), args.Error(1)
}

func mockFactory(mock orchestration.NodeAPI) func(c contractapi.TransactionContextInterface) (orchestration.NodeAPI, error) {
	return func(_ contractapi.TransactionContextInterface) (orchestration.NodeAPI, error) {
		return mock, nil
	}
}

func TestRegistration(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	org := "TestOrg"

	o := &assets.Node{Id: org}
	mockService.On("RegisterNode", o).Return(nil).Once()

	sID := msp.SerializedIdentity{
		Mspid: org,
	}
	b, err := proto.Marshal(&sID)
	require.Nil(t, err, "SID marshal should not fail")

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(testHelper.MockedContext)
	ctx.On("GetStub").Return(stub).Once()

	node, err := contract.RegisterNode(ctx)
	assert.Nil(t, err, "node registration should not fail")

	assert.Equal(t, node, o)
}

func TestQueryNodes(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	nodes := []*assets.Node{
		{Id: "org1"},
		{Id: "org2"},
	}

	mockService.On("GetNodes").Return(nodes, nil).Once()

	ctx := new(testHelper.MockedContext)
	resp, err := contract.QueryNodes(ctx)
	assert.Nil(t, err, "querying nodes should not fail")
	assert.Len(t, resp, len(nodes), "query should return all nodes")
}
