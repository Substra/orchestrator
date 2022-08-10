package adapters

import (
	"context"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/server/distributed/interceptors"
)

// OrganizationAdapter is a grpc server exposing the same organization interface,
// but relies on a remote chaincode to actually manage the asset.
type OrganizationAdapter struct {
	asset.UnimplementedOrganizationServiceServer
}

// NewOrganizationAdapter creates a Server
func NewOrganizationAdapter() *OrganizationAdapter {
	return &OrganizationAdapter{}
}

// RegisterOrganization will add a new organization to the network
func (a *OrganizationAdapter) RegisterOrganization(ctx context.Context, in *asset.RegisterOrganizationParam) (*asset.Organization, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.organization:RegisterOrganization"

	organization := &asset.Organization{}

	err = invocator.Call(ctx, method, in, organization)

	return organization, err
}

// GetAllOrganizations will return all known organizations
func (a *OrganizationAdapter) GetAllOrganizations(ctx context.Context, in *asset.GetAllOrganizationsParam) (*asset.GetAllOrganizationsResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.organization:GetAllOrganizations"

	organizations := &asset.GetAllOrganizationsResponse{}

	err = invocator.Call(ctx, method, in, organizations)

	return organizations, err
}
