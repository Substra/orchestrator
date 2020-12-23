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

type ObjectiveResponse struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	TestDataset string            `json:"testDataset"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
}

func responseFromAsset(o *objectiveAsset.Objective) *ObjectiveResponse {
	return &ObjectiveResponse{
		Key:         o.Key,
		Name:        o.Name,
		TestDataset: o.TestDataset,
		Permissions: o.Permissions,
		Metadata:    o.Metadata,
	}
}

func (s *SmartContract) QueryObjective(ctx contractapi.TransactionContextInterface, key string) (*ObjectiveResponse, error) {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return nil, err
	}

	o, err := service.GetObjective(key)
	return responseFromAsset(o), err
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryObjective"}
}
