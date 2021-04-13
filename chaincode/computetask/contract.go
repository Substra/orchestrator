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

package computetask

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// SmartContract manages ComputeTask
type SmartContract struct {
	contractapi.Contract
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "org.substra.computetask"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.AfterTransaction = ledger.AfterTransactionHook

	return contract
}

func (s *SmartContract) RegisterTask(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetComputeTaskService()

	newTask := new(asset.NewComputeTask)
	err := wrapper.Unwrap(newTask)
	if err != nil {
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	t, err := service.RegisterTask(newTask, owner)
	if err != nil {
		return nil, err
	}
	wrapped, err := communication.Wrap(t)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) QueryTasks(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetComputeTaskService()

	param := new(asset.QueryTasksParam)
	err := wrapper.Unwrap(param)
	if err != nil {
		return nil, err
	}

	pagination := common.NewPagination(param.PageToken, param.PageSize)

	tasks, nextPage, err := service.GetTasks(pagination, param.Filter)
	if err != nil {
		return nil, err
	}

	resp := &asset.QueryTasksResponse{
		Tasks:         tasks,
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
	return []string{"QueryTasks"}
}
