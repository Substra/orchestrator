package computeplan

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

// SmartContract manages ComputePlan
type SmartContract struct {
	contractapi.Contract
	logger zerolog.Logger
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.computeplan"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.With().Str("contract", contract.Name).Logger()

	return contract
}

func (s *SmartContract) RegisterPlan(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputePlanService()

	newPlan := new(asset.NewComputePlan)
	err = wrapper.Unwrap(newPlan)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to extract tx creator")
		return nil, err
	}

	t, err := service.RegisterPlan(newPlan, owner)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to register compute plan")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), t)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) GetPlan(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputePlanService()

	param := new(asset.GetComputePlanParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	t, err := service.GetPlan(param.Key)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get compute plan")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), t)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) QueryPlans(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputePlanService()

	param := new(asset.QueryPlansParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	plans, nextPage, err := service.QueryPlans(common.NewPagination(param.PageToken, param.PageSize), param.Filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query compute plans")
		return nil, err
	}
	resp := &asset.QueryPlansResponse{
		Plans:         plans,
		NextPageToken: nextPage,
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) ApplyPlanAction(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	provider, err := ctx.GetProvider()
	if err != nil {
		return err
	}
	service := provider.GetComputePlanService()

	param := new(asset.ApplyPlanActionParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to extract tx creator")
		return err
	}

	err = service.ApplyPlanAction(param.Key, param.Action, requester)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to apply compute plan action")
		return err
	}

	return nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["ComputePlan"]
}

// UpdatePlan updates a Compute Plan in world state
// If the key does not exist, it will throw an error
func (s *SmartContract) UpdatePlan(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	provider, err := ctx.GetProvider()
	if err != nil {
		return err
	}
	service := provider.GetComputePlanService()

	params := new(asset.UpdateComputePlanParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to extract tx creator")
		return err
	}

	err = service.UpdatePlan(params, requester)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to update Compute Plan")
		return err
	}

	return nil
}

func (s *SmartContract) IsPlanRunning(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetComputePlanService()

	param := new(asset.IsPlanRunningParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	isRunning, err := service.IsPlanRunning(param.Key)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get compute plan")
		return nil, err
	}

	resp := &asset.IsPlanRunningResponse{IsRunning: isRunning}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}
