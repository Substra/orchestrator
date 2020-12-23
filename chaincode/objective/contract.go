package objective

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substrafoundation/substra-orchestrator/chaincode/ledger"
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

func NewSmartContract() *SmartContract {
	return &SmartContract{
		serviceFactory: getServiceFromContext,
	}
}

// RegisterObjective creates a new objective in world state
func (s *SmartContract) RegisterObjective(ctx contractapi.TransactionContextInterface, id string) error {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return err
	}

	o := objectiveAsset.Objective{Key: id}

	err = service.RegisterObjective(&o)
	return err
}

func (s *SmartContract) QueryObjective(ctx contractapi.TransactionContextInterface, key string) (*objectiveAsset.Objective, error) {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return nil, err
	}

	return service.GetObjective(key)
}
