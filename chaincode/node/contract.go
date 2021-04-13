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
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
)

// SmartContract manages nodes
type SmartContract struct {
	contractapi.Contract
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "org.substra.node"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.AfterTransaction = ledger.AfterTransactionHook

	return contract
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryNodes"}
}

// RegisterNode creates a new node in world state
func (s *SmartContract) RegisterNode(ctx ledger.TransactionContext) (*communication.Wrapper, error) {
	txCreator, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	service := ctx.GetProvider().GetNodeService()

	node, err := service.RegisterNode(txCreator)
	if err != nil {
		return nil, err
	}
	wrapped, err := communication.Wrap(node)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

// QueryNodes retrieves all known nodes
func (s *SmartContract) QueryNodes(ctx ledger.TransactionContext) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetNodeService()

	nodes, err := service.GetNodes()
	if err != nil {
		return nil, err
	}

	resp := &asset.NodeQueryResponse{
		Nodes: nodes,
	}

	wrapped, err := communication.Wrap(resp)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}
