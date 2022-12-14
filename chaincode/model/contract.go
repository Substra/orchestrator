package model

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	"github.com/substra/orchestrator/lib/asset"
	commonserv "github.com/substra/orchestrator/server/common"
)

// SmartContract manages Models
type SmartContract struct {
	contractapi.Contract
	logger zerolog.Logger
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.model"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.With().Str("contract", contract.Name).Logger()

	return contract
}

// RegisterModel associates a new model to a running task
func (s *SmartContract) RegisterModel(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetModelService()

	params := new(asset.NewModel)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to extract tx creator")
		return nil, err
	}

	obj, err := service.RegisterModels([]*asset.NewModel{params}, requester)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to register model")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), obj[0])
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) GetModel(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetModelService()

	params := new(asset.GetModelParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap params")
		return nil, err
	}

	model, err := service.GetModel(params.GetKey())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to fetch model")
		return nil, err
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), model)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) GetComputeTaskOutputModels(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetModelService()

	param := new(asset.GetComputeTaskModelsParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	models, err := service.GetComputeTaskOutputModels(param.ComputeTaskKey)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get models for compute task")
		return nil, err
	}
	response := &asset.GetComputeTaskModelsResponse{
		Models: models,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), response)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) RegisterModels(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetModelService()

	params := new(asset.RegisterModelsParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to extract tx creator")
		return nil, err
	}

	models, err := service.RegisterModels(params.Models, owner)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to register models")
		return nil, err
	}

	resp := &asset.RegisterModelsResponse{
		Models: models,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Model"]
}
