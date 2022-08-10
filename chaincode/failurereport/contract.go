package failurereport

import (
	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	"github.com/substra/orchestrator/lib/asset"
	commonserv "github.com/substra/orchestrator/server/common"
)

// SmartContract manages FailureReports
type SmartContract struct {
	contractapi.Contract
	logger log.Entry
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.failurereport"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

// RegisterFailureReport associates a new failure report to a failed task
func (s *SmartContract) RegisterFailureReport(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetFailureReportService()

	params := new(asset.NewFailureReport)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap params")
		return nil, err
	}

	requester, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return nil, err
	}

	obj, err := service.RegisterFailureReport(params, requester)
	if err != nil {
		s.logger.WithError(err).Error("failed to register failure report")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), obj)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

func (s *SmartContract) GetFailureReport(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetFailureReportService()

	params := new(asset.GetFailureReportParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap params")
		return nil, err
	}

	model, err := service.GetFailureReport(params.GetComputeTaskKey())
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch failure report")
		return nil, err
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), model)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["FailureReport"]
}
