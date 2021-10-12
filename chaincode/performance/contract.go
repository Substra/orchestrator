// Package performance containce the smartcontract related to training performance management.
package performance

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
	contract.Name = "orchestrator.performance"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Performance"]
}

func (s *SmartContract) RegisterPerformance(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	ctx.SetRequestID(wrapper.RequestID)
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetPerformanceService()

	newPerf := new(asset.NewPerformance)
	err = wrapper.Unwrap(newPerf)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	t, err := service.RegisterPerformance(newPerf, requester)
	if err != nil {
		s.logger.WithError(err).Error("failed to register performance")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), t)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) QueryPerformances(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	ctx.SetRequestID(wrapper.RequestID)
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetPerformanceService()

	param := new(asset.QueryPerformancesParam)
	err = wrapper.Unwrap(param)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	performances, nextPage, err := service.QueryPerformances(&common.Pagination{Token: param.GetPageToken(), Size: param.GetPageSize()}, param.Filter)
	if err != nil {
		s.logger.WithError(err).Error("failed to query performances")
		return nil, err
	}

	resp := &asset.QueryPerformancesResponse{
		Performances:  performances,
		NextPageToken: nextPage,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}
