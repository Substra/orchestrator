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
	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	commonserv "github.com/owkin/orchestrator/server/common"
)

// SmartContract manages ComputeTask
type SmartContract struct {
	contractapi.Contract
	logger log.Entry
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.computetask"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

func (s *SmartContract) RegisterTask(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetComputeTaskService()

	newTask := new(asset.NewComputeTask)
	err := wrapper.Unwrap(newTask)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	t, err := service.RegisterTask(newTask, owner)
	if err != nil {
		s.logger.WithError(err).Error("failed to register task")
		return nil, err
	}
	wrapped, err := communication.Wrap(t)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) RegisterTasks(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetComputeTaskService()

	newTasks := new(asset.RegisterTasksParam)
	err := wrapper.Unwrap(newTasks)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	tasks, err := service.RegisterTasks(newTasks.GetTasks(), owner)
	if err != nil {
		s.logger.WithError(err).Error("failed to register computetasks")
		return nil, err
	}

	resp := &asset.RegisterTasksResponse{
		Tasks: tasks,
	}

	wrapped, err := communication.Wrap(resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, err
}

// GetTask returns the task with given key
func (s *SmartContract) GetTask(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetComputeTaskService()

	params := new(asset.GetTaskParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	task, err := service.GetTask(params.GetKey())
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch computetask")
		return nil, err
	}

	wrapped, err := communication.Wrap(task)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) QueryTasks(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetComputeTaskService()

	param := new(asset.QueryTasksParam)
	err := wrapper.Unwrap(param)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	pagination := common.NewPagination(param.PageToken, param.PageSize)

	tasks, nextPage, err := service.QueryTasks(pagination, param.Filter)
	if err != nil {
		s.logger.WithError(err).Error("failed to query tasks")
		return nil, err
	}

	resp := &asset.QueryTasksResponse{
		Tasks:         tasks,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) ApplyTaskAction(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	service := ctx.GetProvider().GetComputeTaskService()

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return err
	}

	param := new(asset.ApplyTaskActionParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return err
	}

	err = service.ApplyTaskAction(param.ComputeTaskKey, param.Action, param.Log, requester)
	if err != nil {
		s.logger.WithError(err).Error("failed to apply task action")
		return err
	}

	return nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["ComputeTask"]
}
