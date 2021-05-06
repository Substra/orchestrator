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

package objective

import (
	"errors"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// SmartContract manages objectives
type SmartContract struct {
	contractapi.Contract
	logger log.Entry
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.objective"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

// RegisterObjective creates a new objective in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterObjective(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetObjectiveService()

	params := new(asset.NewObjective)
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

	obj, err := service.RegisterObjective(params, owner)
	if err != nil {
		s.logger.WithError(err).Error("failed to register objective")
		return nil, err
	}
	wrapped, err := communication.Wrap(obj)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// GetObjective returns the objective with given key
func (s *SmartContract) GetObjective(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetObjectiveService()

	params := new(asset.GetObjectiveParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	obj, err := service.GetObjective(params.GetKey())
	if err != nil {
		s.logger.WithError(err).Error("failed to query objective")
		return nil, err
	}
	wrapped, err := communication.Wrap(obj)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryObjectives returns the objectives
func (s *SmartContract) QueryObjectives(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetObjectiveService()

	params := new(asset.QueryObjectivesParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	objectives, nextPage, err := service.QueryObjectives(&common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()})
	if err != nil {
		s.logger.WithError(err).Error("failed to query objectives")
		return nil, err
	}

	resp := &asset.QueryObjectivesResponse{
		Objectives:    objectives,
		NextPageToken: nextPage,
	}
	wrapped, err := communication.Wrap(resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryLeaderboard returns for an objective all its certified testtuples with a done status
func (s *SmartContract) QueryLeaderboard(ctx ledger.TransactionContext, key string, sortOrder asset.SortOrder) (*asset.Leaderboard, error) {
	return nil, errors.New("unimplemented")
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"GetObjective", "QueryObjectives", "QueryLeaderboard"}
}
