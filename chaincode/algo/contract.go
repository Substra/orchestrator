package algo

import (
	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	commonserv "github.com/substra/orchestrator/server/common"
)

// SmartContract manages algos
type SmartContract struct {
	contractapi.Contract
	logger log.Entry
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.algo"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

// RegisterAlgo creates a new algo in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterAlgo(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetAlgoService()

	params := new(asset.NewAlgo)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	a, err := service.RegisterAlgo(params, owner)
	if err != nil {
		s.logger.WithError(err).Error("failed to register algo")
		return nil, err
	}
	response, err := communication.Wrap(ctx.GetContext(), a)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return response, nil
}

// GetAlgo returns the algo with given key
func (s *SmartContract) GetAlgo(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetAlgoService()

	params := new(asset.GetAlgoParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	algo, err := service.GetAlgo(params.GetKey())
	if err != nil {
		s.logger.WithError(err).Error("failed to query algo")
		return nil, err
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), algo)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryAlgos returns the algos
func (s *SmartContract) QueryAlgos(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetAlgoService()

	params := new(asset.QueryAlgosParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	algos, nextPage, err := service.QueryAlgos(&common.Pagination{Token: params.GetPageToken(), Size: params.GetPageSize()}, params.Filter)
	if err != nil {
		s.logger.WithError(err).Error("failed to query algos")
		return nil, err
	}

	resp := &asset.QueryAlgosResponse{
		Algos:         algos,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// UpdateAlgo updates an algo in world state
// If the key does not exist, it will throw an error
func (s *SmartContract) UpdateAlgo(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	provider, err := ctx.GetProvider()
	if err != nil {
		return err
	}
	service := provider.GetAlgoService()

	params := new(asset.UpdateAlgoParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return err
	}

	err = service.UpdateAlgo(params, requester)
	if err != nil {
		s.logger.WithError(err).Error("failed to update algo")
		return err
	}

	return nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Algo"]
}
