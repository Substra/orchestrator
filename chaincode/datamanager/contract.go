package datamanager

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

// SmartContract manages datasamples
type SmartContract struct {
	contractapi.Contract
	logger zerolog.Logger
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.datamanager"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.With().Str("contract", contract.Name).Logger()

	return contract
}

// RegisterDataManager creates a new data Manager in world state
// If the key exists, it will throw an error
func (s *SmartContract) RegisterDataManager(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetDataManagerService()

	params := new(asset.NewDataManager)
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

	dm, err := service.RegisterDataManager(params, owner)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to register datamanager")
		return nil, err
	}

	response, err := communication.Wrap(ctx.GetContext(), dm)

	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}

	return response, nil
}

// GetDataManager returns the DataManager with given key
func (s *SmartContract) GetDataManager(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetDataManagerService()

	params := new(asset.GetDataManagerParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	dataManager, err := service.GetDataManager(params.GetKey())
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query datamanager")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), dataManager)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryDataManagers returns the DataManager
func (s *SmartContract) QueryDataManagers(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetDataManagerService()

	params := new(asset.QueryDataManagersParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to unwrap param")
		return nil, err
	}

	datamanagers, nextPage, err := service.QueryDataManagers(&common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query datamanagers")
		return nil, err
	}

	resp := &asset.QueryDataManagersResponse{
		DataManagers:  datamanagers,
		NextPageToken: nextPage,
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
	return commonserv.ReadOnlyMethods["DataManager"]
}

// UpdateDataManager updates an DataManager in world state
// If the key does not exist, it will throw an error
func (s *SmartContract) UpdateDataManager(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	provider, err := ctx.GetProvider()
	if err != nil {
		return err
	}
	service := provider.GetDataManagerService()

	params := new(asset.UpdateDataManagerParam)
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

	err = service.UpdateDataManager(params, requester)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to update DataManager")
		return err
	}

	return nil
}
