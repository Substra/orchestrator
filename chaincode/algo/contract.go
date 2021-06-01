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
	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	commonserv "github.com/owkin/orchestrator/server/common"
)

// SmartContract manages algos
type SmartContract struct {
	contractapi.Contract
	logger log.Entry
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.algo"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

// RegisterAlgo creates a new algo in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterAlgo(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetAlgoService()

	params := new(asset.NewAlgo)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	a, err := service.RegisterAlgo(params, owner)
	if err != nil {
		s.logger.WithError(err).Error("failed to register algo")
		return nil, err
	}
	response, err := communication.Wrap(a)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return response, nil
}

// GetAlgo returns the algo with given key
func (s *SmartContract) GetAlgo(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetAlgoService()

	params := new(asset.GetAlgoParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	algo, err := service.GetAlgo(params.GetKey())
	if err != nil {
		s.logger.WithError(err).Error("failed to query algo")
		return nil, err
	}

	wrapped, err := communication.Wrap(algo)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryAlgos returns the algos
func (s *SmartContract) QueryAlgos(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetAlgoService()

	params := new(asset.QueryAlgosParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	algos, nextPage, err := service.QueryAlgos(params.Category, &common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()})
	if err != nil {
		s.logger.WithError(err).Error("failed to query algos")
		return nil, err
	}

	resp := &asset.QueryAlgosResponse{
		Algos:         algos,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Algo"]
}
