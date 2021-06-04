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

package model

import (
	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	commonserv "github.com/owkin/orchestrator/server/common"
)

// SmartContract manages Models
type SmartContract struct {
	contractapi.Contract
	logger log.Entry
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.model"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

// RegisterModel associates a new model to a running task
func (s *SmartContract) RegisterModel(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetModelService()

	params := new(asset.NewModel)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	obj, err := service.RegisterModel(params, requester)
	if err != nil {
		s.logger.WithError(err).Error("failed to register model")
		return nil, err
	}
	wrapped, err := communication.Wrap(obj)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) GetModel(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetModelService()

	params := new(asset.GetModelParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap params")
		return nil, err
	}

	model, err := service.GetModel(params.GetKey())
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch model")
		return nil, err
	}

	wrapped, err := communication.Wrap(model)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryModels returns the models
func (s *SmartContract) QueryModels(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetModelService()

	params := new(asset.QueryModelsParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	models, nextPage, err := service.QueryModels(params.Category, common.NewPagination(params.GetPageToken(), params.GetPageSize()))
	if err != nil {
		s.logger.WithError(err).Error("failed to query models")
		return nil, err
	}

	resp := &asset.QueryModelsResponse{
		Models:        models,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) GetComputeTaskOutputModels(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetModelService()

	param := new(asset.GetComputeTaskModelsParam)
	err := wrapper.Unwrap(param)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	models, err := service.GetComputeTaskOutputModels(param.ComputeTaskKey)
	if err != nil {
		s.logger.WithError(err).Error("failed to get models for compute task")
		return nil, err
	}
	response := &asset.GetComputeTaskModelsResponse{
		Models: models,
	}

	wrapped, err := communication.Wrap(response)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) GetComputeTaskInputModels(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetModelService()

	param := new(asset.GetComputeTaskModelsParam)
	err := wrapper.Unwrap(param)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	models, err := service.GetComputeTaskInputModels(param.ComputeTaskKey)
	if err != nil {
		s.logger.WithError(err).Error("failed to get input models for compute task")
		return nil, err
	}
	response := &asset.GetComputeTaskModelsResponse{
		Models: models,
	}

	wrapped, err := communication.Wrap(response)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) CanDisableModel(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	service := ctx.GetProvider().GetModelService()

	params := new(asset.CanDisableModelParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	can, err := service.CanDisableModel(params.ModelKey, requester)
	if err != nil {
		s.logger.WithError(err).Error("failed to check whether model can be disabled")
		return nil, err
	}
	wrapped, err := communication.Wrap(&asset.CanDisableModelResponse{CanDisable: can})
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) DisableModel(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	service := ctx.GetProvider().GetModelService()

	params := new(asset.DisableModelParam)
	err := wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return err
	}

	err = service.DisableModel(params.ModelKey, requester)
	if err != nil {
		s.logger.WithError(err).Error("failed to check whether model can be disabled")
		return err
	}

	return nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Model"]
}
