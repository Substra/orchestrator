package organization

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	"github.com/substra/orchestrator/lib/asset"
	commonserv "github.com/substra/orchestrator/server/common"
)

// SmartContract manages organizations
type SmartContract struct {
	contractapi.Contract
	logger zerolog.Logger
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.organization"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.With().Str("contract", contract.Name).Logger()

	return contract
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Organization"]
}

// RegisterOrganization creates a new organization in world state
func (s *SmartContract) RegisterOrganization(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	txCreator, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to extract tx creator")
		return nil, err
	}

	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetOrganizationService()

	params := new(asset.RegisterOrganizationParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	organization, err := service.RegisterOrganization(txCreator, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to register organization")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), organization)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// GetAllOrganizations retrieves all known organizations
func (s *SmartContract) GetAllOrganizations(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetOrganizationService()

	organizations, err := service.GetAllOrganizations()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query organizations")
		return nil, err
	}

	resp := &asset.GetAllOrganizationsResponse{
		Organizations: organizations,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}
