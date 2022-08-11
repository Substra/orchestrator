package info

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	"github.com/substra/orchestrator/lib/asset"
	commonserv "github.com/substra/orchestrator/server/common"
)

// Version represents the version of the chaincode, the value is changed at build time
var Version = "dev"

// SmartContract manages info
type SmartContract struct {
	contractapi.Contract
	logger zerolog.Logger
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	contract := &SmartContract{}
	contract.Name = "orchestrator.info"
	contract.TransactionContextHandler = ledger.NewContext()
	contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
	contract.AfterTransaction = ledger.AfterTransactionHook

	contract.logger = log.With().Str("contract", contract.Name).Logger()

	return contract
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return commonserv.ReadOnlyMethods["Info"]
}

// QueryVersion returns the chaincode version
func (s *SmartContract) QueryVersion(ctx ledger.TransactionContext, wrapper *communication.Wrapper) (*communication.Wrapper, error) {
	info := &asset.QueryVersionResponse{
		Chaincode: Version,
	}

	wrapped, err := communication.Wrap(ctx.GetContext(), info)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to wrap response")
		return nil, err
	}
	return wrapped, nil
}
