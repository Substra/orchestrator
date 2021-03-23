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

package datamanager

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// SmartContract manages datasamples
type SmartContract struct {
	contractapi.Contract
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "org.substra.datamanager"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.AfterTransaction = ledger.AfterTransactionHook

	return contract
}

// RegisterDataManager creates a new data Manager in world state
// If the key exists, it will throw an error
func (s *SmartContract) RegisterDataManager(ctx ledger.TransactionContext, params *asset.NewDataManager) (*asset.NewDataManagerResponse, error) {
	service := ctx.GetProvider().GetDataManagerService()

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	err = service.RegisterDataManager(params, owner)
	if err != nil {
		return nil, err
	}
	return &asset.NewDataManagerResponse{}, nil
}

// UpdateDataManager updates a data manager in world state
// If the key does not exist, it will throw an error
func (s *SmartContract) UpdateDataManager(ctx ledger.TransactionContext, params *asset.DataManagerUpdateParam) (*asset.DataManagerUpdateResponse, error) {
	service := ctx.GetProvider().GetDataManagerService()

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	err = service.UpdateDataManager(params, owner)
	if err != nil {
		return nil, err
	}
	return &asset.DataManagerUpdateResponse{}, nil
}

// QueryDataManager returns the DataManager with given key
func (s *SmartContract) QueryDataManager(ctx ledger.TransactionContext, params *asset.DataManagerQueryParam) (*asset.DataManager, error) {
	service := ctx.GetProvider().GetDataManagerService()

	return service.GetDataManager(params.GetKey())
}

// QueryDataManagers returns the DataManager
func (s *SmartContract) QueryDataManagers(ctx ledger.TransactionContext, params *asset.DataManagersQueryParam) (*asset.DataManagersQueryResponse, error) {
	service := ctx.GetProvider().GetDataManagerService()

	datamanagers, nextPage, err := service.GetDataManagers(&common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()})
	if err != nil {
		return nil, err
	}

	return &asset.DataManagersQueryResponse{
		DataManagers:  datamanagers,
		NextPageToken: nextPage,
	}, nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryDataManager", "QueryDataManagers"}
}
