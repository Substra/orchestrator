package objective

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/mocks"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *mocks.TransactionContext) *service.MockObjectiveAPI {
	mockService := new(service.MockObjectiveAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetObjectiveService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()
	ctx.On("SetRequestID", "").Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	addressable := &asset.Addressable{}
	newPerms := &asset.NewPermissions{}
	metadata := map[string]string{"test": "true"}

	mspid := "org"

	newObj := &asset.NewObjective{
		Key:            "uuid1",
		Name:           "Objective name",
		Description:    addressable,
		MetricsName:    "metrics name",
		Metrics:        addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
	}
	wrapper, err := communication.Wrap(context.Background(), newObj)
	assert.NoError(t, err)

	o := &asset.Objective{}

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterObjective", newObj, mspid).Return(o, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterObjective(ctx, wrapper)
	assert.NoError(t, err)
}

func TestQueryObjectives(t *testing.T) {
	contract := &SmartContract{}

	objectives := []*asset.Objective{
		{Name: "test"},
		{Name: "test2"},
	}

	ctx := new(mocks.TransactionContext)
	service := getMockedService(ctx)
	service.On("QueryObjectives", &common.Pagination{Token: "", Size: 20}).Return(objectives, "", nil).Once()

	param := &asset.QueryObjectivesParam{PageToken: "", PageSize: 20}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	wrapped, err := contract.QueryObjectives(ctx, wrapper)
	assert.NoError(t, err, "query should not fail")

	resp := new(asset.QueryObjectivesResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Objectives, len(objectives), "query should return all objectives")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"GetObjective",
		"QueryObjectives",
		"GetLeaderboard",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
