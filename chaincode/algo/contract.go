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

package algo

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// SmartContract manages algos
type SmartContract struct {
	contractapi.Contract
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "org.substra.algo"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	return contract
}

// RegisterAlgo creates a new algo in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterAlgo(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetAlgoService()

	params := new(asset.NewAlgo)
	err := wrapper.Unwrap(params)
	if err != nil {
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	a, err := service.RegisterAlgo(params, owner)
	if err != nil {
		return nil, err
	}
	response, err := communication.Wrap(a)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// QueryAlgo returns the algo with given key
func (s *SmartContract) QueryAlgo(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetAlgoService()

	params := new(asset.AlgoQueryParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		return nil, err
	}

	algo, err := service.GetAlgo(params.GetKey())
	if err != nil {
		return nil, err
	}

	wrapped, err := communication.Wrap(algo)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

// QueryAlgos returns the algos
func (s *SmartContract) QueryAlgos(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetAlgoService()

	params := new(asset.AlgosQueryParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		return nil, err
	}

	algos, nextPage, err := service.GetAlgos(params.Category, &common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()})
	if err != nil {
		return nil, err
	}

	resp := &asset.AlgosQueryResponse{
		Algos:         algos,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(resp)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryAlgo", "QueryAlgos"}
}
