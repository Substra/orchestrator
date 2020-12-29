package objective

import (
	"errors"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substrafoundation/substra-orchestrator/chaincode/ledger"
	"github.com/substrafoundation/substra-orchestrator/lib/assets"
	objectiveAsset "github.com/substrafoundation/substra-orchestrator/lib/assets/objective"
)

func getServiceFromContext(ctx contractapi.TransactionContextInterface) (objectiveAsset.API, error) {
	db, err := ledger.GetLedgerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return objectiveAsset.NewService(db), nil
}

// SmartContract manages objectives
type SmartContract struct {
	contractapi.Contract
	serviceFactory func(contractapi.TransactionContextInterface) (objectiveAsset.API, error)
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	return &SmartContract{
		serviceFactory: getServiceFromContext,
	}
}

// RegisterObjective creates a new objective in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterObjective(ctx contractapi.TransactionContextInterface, id string) error {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return err
	}

	o := objectiveAsset.Objective{Key: id}

	err = service.RegisterObjective(&o)
	return err
}

// QueryObjectives returns the objectives
func (s *SmartContract) QueryObjectives(ctx contractapi.TransactionContextInterface) ([]*objectiveAsset.Objective, error) {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return nil, err
	}

	return service.GetObjectives()
}

// QueryLeaderboard returns for an objective all its certified testtuples with a done status
func (s *SmartContract) QueryLeaderboard(ctx contractapi.TransactionContextInterface, key string, sortOrder assets.SortOrder) (*objectiveAsset.Leaderboard, error) {
	return nil, errors.New("unimplemented")
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryObjectives", "QueryLeaderboard"}
}
