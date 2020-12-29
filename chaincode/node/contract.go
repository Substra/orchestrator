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
	"github.com/substrafoundation/substra-orchestrator/chaincode/ledger"
	nodeAsset "github.com/substrafoundation/substra-orchestrator/lib/assets/node"
)

func getServiceFromContext(ctx contractapi.TransactionContextInterface) (nodeAsset.API, error) {
	db, err := ledger.GetLedgerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return nodeAsset.NewService(db), nil
}

// SmartContract manages nodes
type SmartContract struct {
	contractapi.Contract
	serviceFactory func(contractapi.TransactionContextInterface) (nodeAsset.API, error)
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	return &SmartContract{
		serviceFactory: getServiceFromContext,
	}
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryNodes"}
}

// RegisterNode creates a new node in world state
func (s *SmartContract) RegisterNode(ctx contractapi.TransactionContextInterface) (*nodeAsset.Node, error) {
	txCreator, err := ledger.GetTxCreator(ctx.GetStub())

	service, err := s.serviceFactory(ctx)
	if err != nil {
		return nil, err
	}

	node := nodeAsset.Node{Id: txCreator}

	err = service.RegisterNode(&node)
	return &node, err
}

// QueryNodes retrieves all known nodes
func (s *SmartContract) QueryNodes(ctx contractapi.TransactionContextInterface) ([]*nodeAsset.Node, error) {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return nil, err
	}

	return service.GetNodes()
}
