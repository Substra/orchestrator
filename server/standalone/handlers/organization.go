package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	commonInterceptors "github.com/owkin/orchestrator/server/common/interceptors"
	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// OrganizationServer is the gRPC server exposing organization actions
type OrganizationServer struct {
	asset.UnimplementedOrganizationServiceServer
}

// NewOrganizationServer creates a Server
func NewOrganizationServer() *OrganizationServer {
	return &OrganizationServer{}
}

// RegisterOrganization will add a new organization to the network
func (s *OrganizationServer) RegisterOrganization(ctx context.Context, in *asset.RegisterOrganizationParam) (*asset.Organization, error) {
	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	organization, err := services.GetOrganizationService().RegisterOrganization(mspid, in)
	if err != nil {
		return nil, err
	}
	return organization, nil
}

// GetAllOrganizations will return all known organizations
func (s *OrganizationServer) GetAllOrganizations(ctx context.Context, in *asset.GetAllOrganizationsParam) (*asset.GetAllOrganizationsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	organizations, err := services.GetOrganizationService().GetAllOrganizations()

	return &asset.GetAllOrganizationsResponse{
		Organizations: organizations,
	}, err
}
