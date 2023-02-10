package function

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

// SmartContract manages functions
type SmartContract struct {
	contractapi.Contract
	logger zerolog.Logger
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.function"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.With().Str("contract", contract.Name).Logger()

	return contract
}

// RegisterFunction creates a new function in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterFunction(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetFunctionService()

	params := new(asset.NewFunction)
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

	a, err := service.RegisterFunction(params, owner)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to register function")
		return nil, err
	}
	response, err := communication.Wrap(ctx.GetContext(), a)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return response, nil
}

// GetFunction returns the function with given key
func (s *SmartContract) GetFunction(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetFunctionService()

	params := new(asset.GetFunctionParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	function, err := service.GetFunction(params.GetKey())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query function")
		return nil, err
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), function)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryFunctions returns the functions
func (s *SmartContract) QueryFunctions(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetFunctionService()

	params := new(asset.QueryFunctionsParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	functions, nextPage, err := service.QueryFunctions(&common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()}, params.Filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query functions")
		return nil, err
	}

	resp := &asset.QueryFunctionsResponse{
		Functions:     functions,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// UpdateFunction updates an function in world state
// If the key does not exist, it will throw an error
func (s *SmartContract) UpdateFunction(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	provider, err := ctx.GetProvider()
	if err != nil {
		return err
	}
	service := provider.GetFunctionService()

	params := new(asset.UpdateFunctionParam)
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

	err = service.UpdateFunction(params, requester)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to update function")
		return err
	}

	return nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Function"]
}
