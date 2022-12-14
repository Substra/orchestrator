package computetask

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	commonserv "github.com/substra/orchestrator/server/common"
)

// SmartContract manages ComputeTask
type SmartContract struct {
	contractapi.Contract
	logger zerolog.Logger
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.computetask"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.With().Str("contract", contract.Name).Logger()

	return contract
}

func (s *SmartContract) RegisterTasks(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputeTaskService()

	newTasks := new(asset.RegisterTasksParam)
	err = wrapper.Unwrap(newTasks)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to extract tx creator")
		return nil, err
	}

	tasks, err := service.RegisterTasks(newTasks.GetTasks(), owner)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to register tasks")
		return nil, err
	}

	resp := &asset.RegisterTasksResponse{
		Tasks: tasks,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// GetTask returns the task with given key
func (s *SmartContract) GetTask(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputeTaskService()

	params := new(asset.GetTaskParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	task, err := service.GetTask(params.GetKey())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to fetch computetask")
		return nil, err
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), task)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) QueryTasks(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputeTaskService()

	param := new(asset.QueryTasksParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	pagination := common.NewPagination(param.PageToken, param.PageSize)

	tasks, nextPage, err := service.QueryTasks(pagination, param.Filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query tasks")
		return nil, err
	}

	resp := &asset.QueryTasksResponse{
		Tasks:         tasks,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) ApplyTaskAction(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	provider, err := ctx.GetProvider()
	if err != nil {
		return err
	}
	service := provider.GetComputeTaskService()

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return err
	}

	param := new(asset.ApplyTaskActionParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return err
	}

	err = service.ApplyTaskAction(param.ComputeTaskKey, param.Action, param.Log, requester)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to apply task action")
		return err
	}

	return nil
}

func (s *SmartContract) GetTaskInputAssets(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputeTaskService()

	params := new(asset.GetTaskInputAssetsParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	inputs, err := service.GetInputAssets(params.ComputeTaskKey)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get task input assets")
		return nil, err
	}

	resp := &asset.GetTaskInputAssetsResponse{Assets: inputs}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) DisableOutput(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}

	service := provider.GetComputeTaskService()

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		return nil, err
	}

	params := new(asset.DisableOutputParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	err = service.DisableOutput(params.ComputeTaskKey, params.Identifier, requester)
	if err != nil {
		return nil, err
	}

	resp := &asset.DisableOutputResponse{}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["ComputeTask"]
}
