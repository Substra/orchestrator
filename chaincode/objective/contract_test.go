package objective

import (
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	testHelper "github.com/substrafoundation/substra-orchestrator/chaincode/testing"
	"github.com/substrafoundation/substra-orchestrator/lib/assets/objective"
)

type MockedService struct {
	mock.Mock
}

func (m *MockedService) RegisterObjective(o *objective.Objective) error {
	args := m.Called(o)
	return args.Error(0)
}

func (m *MockedService) GetObjective(key string) (*objective.Objective, error) {
	args := m.Called(key)
	return args.Get(0).(*objective.Objective), args.Error(1)
}

func mockFactory(mock objective.API) func(c contractapi.TransactionContextInterface) (objective.API, error) {
	return func(_ contractapi.TransactionContextInterface) (objective.API, error) {
		return mock, nil
	}
}

func TestRegistration(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	o := &objective.Objective{Key: "uuid1"}
	mockService.On("RegisterObjective", o).Return(nil).Once()

	ctx := new(testHelper.MockedContext)
	contract.RegisterObjective(ctx, "uuid1")
}

func TestGetObjective(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	o := &objective.Objective{Key: "uuid"}

	mockService.On("GetObjective", "uuid").Return(o, nil).Once()

	ctx := new(testHelper.MockedContext)
	contract.QueryObjective(ctx, "uuid")
}

func TestResponseFromAsset(t *testing.T) {
	objective := objective.Objective{}

	r := responseFromAsset(&objective)

	assert.Equal(t, objective.Key, r.Key, "Key should be equal")
	assert.Equal(t, objective.Name, r.Name, "Name should be equal")
	assert.Equal(t, objective.TestDataset, r.TestDataset, "TestDataset should be equal")
	assert.Equal(t, objective.Permissions, r.Permissions, "Permissions should be equal")
	assert.Equal(t, objective.Metadata, r.Metadata, "Metadata should be equal")
}
