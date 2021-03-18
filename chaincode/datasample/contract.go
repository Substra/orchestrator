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

package datasample

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
	contract.Name = "org.substra.datasample"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.AfterTransaction = ledger.AfterTransactionHook

	return contract
}

// RegisterDataSample creates a new data sample in world state
// If the key exists, it will throw an error
func (s *SmartContract) RegisterDataSample(ctx ledger.TransactionContext, params *asset.NewDataSample) (*asset.NewDataSampleResponse, error) {
	service := ctx.GetProvider().GetDataSampleService()

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	err = service.RegisterDataSample(params, owner)
	if err != nil {
		return nil, err
	}
	return &asset.NewDataSampleResponse{}, nil
}

// UpdateDataSample updates a data sample in world state
// If the key does not exist, it will throw an error
func (s *SmartContract) UpdateDataSample(ctx ledger.TransactionContext, params *asset.DataSampleUpdateParam) (*asset.DataSampleUpdateResponse, error) {
	service := ctx.GetProvider().GetDataSampleService()

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	err = service.UpdateDataSample(params, owner)
	if err != nil {
		return nil, err
	}
	return &asset.DataSampleUpdateResponse{}, nil
}

// QueryDataSamples returns the datasamples
func (s *SmartContract) QueryDataSamples(ctx ledger.TransactionContext, params *asset.DataSamplesQueryParam) (*asset.DataSamplesQueryResponse, error) {
	service := ctx.GetProvider().GetDataSampleService()

	datasamples, paginationToken, err := service.GetDataSamples(&common.Pagination{Token: params.PageToken, Size: params.GetPageSize()})
	if err != nil {
		return nil, err
	}

	return &asset.DataSamplesQueryResponse{
		DataSamples:   datasamples,
		NextPageToken: paginationToken,
	}, nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryDataSamples"}
}
