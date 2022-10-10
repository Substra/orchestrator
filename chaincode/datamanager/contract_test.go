package datamanager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/service"
)

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	query := []string{"GetDataManager", "QueryDataManagers"}

	assert.Equal(t, query, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}

func getMockedService(ctx *ledger.MockTransactionContext) *service.MockDataManagerAPI {
	mockService := new(service.MockDataManagerAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetDataManagerService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestQueryDataManagers(t *testing.T) {
	contract := &SmartContract{}

	datamanagers := []*asset.DataManager{
		{Key: "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		{Key: "9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
	}

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("QueryDataManagers", &common.Pagination{Token: "", Size: 10}).Return(datamanagers, "", nil).Once()

	param := &asset.QueryDataManagersParam{PageToken: "", PageSize: 10}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	wrapped, err := contract.QueryDataManagers(ctx, wrapper)
	assert.NoError(t, err, "Query should not fail")
	resp := new(asset.QueryDataManagersResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Len(t, resp.DataManagers, len(datamanagers), "Query should return all datasamples")
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	addressable := &asset.Addressable{}
	newPerms := &asset.NewPermissions{}
	metadata := map[string]string{"test": "true"}

	mspid := "org"

	newObj := &asset.NewDataManager{
		Key:            "uuid1",
		Name:           "Datamanager name",
		Description:    addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
		Opener:         addressable,
		Type:           "test",
	}

	wrapper, err := communication.Wrap(context.Background(), newObj)
	assert.NoError(t, err)

	dm := &asset.DataManager{}

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterDataManager", newObj, mspid).Return(dm, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterDataManager(ctx, wrapper)
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	updateDataManagerParam := &asset.UpdateDataManagerParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated data manager name",
	}
	wrapper, err := communication.Wrap(context.Background(), updateDataManagerParam)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("UpdateDataManager", updateDataManagerParam, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.UpdateDataManager(ctx, wrapper)
	assert.NoError(t, err, "Smart contract execution should not fail")
}

func TestArchive(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	archiveDataManagerParam := &asset.ArchiveDataManagerParam{
		Key:      "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Archived: true,
	}
	wrapper, err := communication.Wrap(context.Background(), archiveDataManagerParam)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("ArchiveDataManager", archiveDataManagerParam, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.ArchiveDataManager(ctx, wrapper)
	assert.NoError(t, err, "Smart contract execution should not fail")
}
