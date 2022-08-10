package organization

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/service"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *ledger.MockTransactionContext) *service.MockOrganizationAPI {
	mockService := new(service.MockOrganizationAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetOrganizationService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	o := &asset.Organization{Id: org, Address: "org-1.com"}
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	newOrganization := &asset.RegisterOrganizationParam{
		Address: "org-1.com",
	}
	service.On("RegisterOrganization", org, newOrganization).Return(o, nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	wrapper, err := communication.Wrap(context.Background(), newOrganization)
	require.NoError(t, err)

	wrapped, err := contract.RegisterOrganization(ctx, wrapper)
	assert.NoError(t, err, "organization registration should not fail")
	organization := new(asset.Organization)
	err = wrapped.Unwrap(organization)
	assert.NoError(t, err)
	assert.Equal(t, organization, o)
}

func TestGetAllOrganizations(t *testing.T) {
	contract := &SmartContract{}

	organizations := []*asset.Organization{
		{Id: "org1"},
		{Id: "org2"},
	}

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("GetAllOrganizations").Return(organizations, nil).Once()

	wrapper, err := communication.Wrap(context.Background(), nil)
	require.NoError(t, err)

	wrapped, err := contract.GetAllOrganizations(ctx, wrapper)
	assert.NoError(t, err, "querying organizations should not fail")
	queryResult := new(asset.GetAllOrganizationsResponse)
	err = wrapped.Unwrap(queryResult)
	assert.NoError(t, err)
	assert.Len(t, queryResult.Organizations, len(organizations), "query should return all organizations")
}
