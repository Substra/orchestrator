package datasample

import (
	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	commonserv "github.com/owkin/orchestrator/server/common"
)

// SmartContract manages datasamples
type SmartContract struct {
	contractapi.Contract
	logger log.Entry
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.datasample"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.WithField("contract", contract.Name)

	return contract
}

// RegisterDataSamples register new data samples in world state
// If the key exists, it will throw an error
func (s *SmartContract) RegisterDataSamples(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetDataSampleService()

	params := new(asset.RegisterDataSamplesParam)
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

	datasamples, err := service.RegisterDataSamples(params.Samples, owner)
	if err != nil {
		s.logger.WithError(err).Error("failed to register datasamples")
		return nil, err
	}

	resp := &asset.RegisterDataSamplesResponse{
		DataSamples: datasamples,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// UpdateDataSamples updates a data sample in world state
// If the key does not exist, it will throw an error
func (s *SmartContract) UpdateDataSamples(ctx ledger.TransactionContext, wrapper *communication.Wrapper) error {
	provider, err := ctx.GetProvider()
	if err != nil {
		return err
	}
	service := provider.GetDataSampleService()

	params := new(asset.UpdateDataSamplesParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return err
	}

	owner, err := ledger.GetTxCreator(ctx.GetStub())
	if err != nil {
		s.logger.WithError(err).Error("failed to extract tx creator")
		return err
	}

	err = service.UpdateDataSamples(params, owner)
	if err != nil {
		s.logger.WithError(err).Error("failed to update datasample")
		return err
	}
	return nil
}

// GetDataSample returns the datasample with given key
func (s *SmartContract) GetDataSample(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetDataSampleService()

	params := new(asset.GetDataSampleParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	obj, err := service.GetDataSample(params.GetKey())
	if err != nil {
		s.logger.WithError(err).Error("failed to query datasample")
		return nil, err
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), obj)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}

// QueryDataSamples returns the datasamples
func (s *SmartContract) QueryDataSamples(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	provider, err := ctx.GetProvider()
	if err != nil {
		return nil, err
	}
	service := provider.GetDataSampleService()

	params := new(asset.QueryDataSamplesParam)
	err = wrapper.Unwrap(params)
	if err != nil {
		s.logger.WithError(err).Error("failed to unwrap param")
		return nil, err
	}

	datasamples, paginationToken, err := service.QueryDataSamples(&common.Pagination{Token: params.PageToken, Size: params.GetPageSize()}, params.Filter)
	if err != nil {
		s.logger.WithError(err).Error("failed to query datasamples")
		return nil, err
	}

	resp := &asset.QueryDataSamplesResponse{
		DataSamples:   datasamples,
		NextPageToken: paginationToken,
	}
	wrapped, err := communication.Wrap(ctx.GetContext(), resp)
	if err != nil {
		s.logger.WithError(err).Error("failed to wrap response")
		return nil, err
	}
	return wrapped, err
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["DataSample"]
}
