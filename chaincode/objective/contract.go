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

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// SmartContract manages objectives
type SmartContract struct {
	contractapi.Contract
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "org.substra.objective"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.AfterTransaction = ledger.AfterTransactionHook

	return contract
}

// RegisterObjective creates a new objective in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterObjective(ctx ledger.TransactionContext, params *asset.NewObjective) (*asset.Objective, error) {
	service := ctx.GetProvider().GetObjectiveService()

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	obj, err := service.RegisterObjective(params, owner)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// QueryObjective returns the objective with given key
func (s *SmartContract) QueryObjective(ctx ledger.TransactionContext, params *asset.ObjectiveQueryParam) (*asset.Objective, error) {
	service := ctx.GetProvider().GetObjectiveService()

	return service.GetObjective(params.GetKey())
}

// QueryObjectives returns the objectives
func (s *SmartContract) QueryObjectives(ctx ledger.TransactionContext, params *asset.ObjectivesQueryParam) (*asset.ObjectivesQueryResponse, error) {
	service := ctx.GetProvider().GetObjectiveService()

	objectives, nextPage, err := service.GetObjectives(&common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()})
	if err != nil {
		return nil, err
	}

	return &asset.ObjectivesQueryResponse{
		Objectives:    objectives,
		NextPageToken: nextPage,
	}, nil
}

// QueryLeaderboard returns for an objective all its certified testtuples with a done status
func (s *SmartContract) QueryLeaderboard(ctx ledger.TransactionContext, key string, sortOrder asset.SortOrder) (*asset.Leaderboard, error) {
	return nil, errors.New("unimplemented")
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryObjective", "QueryObjectives", "QueryLeaderboard"}
}
