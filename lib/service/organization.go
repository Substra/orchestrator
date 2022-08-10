package service

import (
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrganizationAPI defines the methods to act on Organizations
type OrganizationAPI interface {
	RegisterOrganization(id string, newOrganization *asset.RegisterOrganizationParam) (*asset.Organization, error)
	GetAllOrganizations() ([]*asset.Organization, error)
	GetOrganization(id string) (*asset.Organization, error)
}

// OrganizationServiceProvider defines an object able to provide a OrganizationAPI instance
type OrganizationServiceProvider interface {
	GetOrganizationService() OrganizationAPI
}

// OrganizationDependencyProvider defines what the OrganizationService needs to perform its duty
type OrganizationDependencyProvider interface {
	persistence.OrganizationDBALProvider
	EventServiceProvider
	TimeServiceProvider
}

// OrganizationService is the organization manipulation entry point
// it implements OrganizationAPI
type OrganizationService struct {
	OrganizationDependencyProvider
}

// NewOrganizationService will create a new service with given persistence layer
func NewOrganizationService(provider OrganizationDependencyProvider) *OrganizationService {
	return &OrganizationService{provider}
}

// RegisterOrganization persist a organization
func (s *OrganizationService) RegisterOrganization(id string, newOrganization *asset.RegisterOrganizationParam) (*asset.Organization, error) {
	err := newOrganization.Validate()
	if err != nil {
		return nil, err
	}

	organization := &asset.Organization{Id: id, Address: newOrganization.Address}

	exists, err := s.GetOrganizationDBAL().OrganizationExists(id)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, orcerrors.NewConflict(asset.OrganizationKind, id)
	}

	organization.CreationDate = timestamppb.New(s.GetTimeService().GetTransactionTime())

	err = s.GetOrganizationDBAL().AddOrganization(organization)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  id,
		AssetKind: asset.AssetKind_ASSET_ORGANIZATION,
		Asset:     &asset.Event_Organization{Organization: organization},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

// GetAllOrganizations list all known organizations
func (s *OrganizationService) GetAllOrganizations() ([]*asset.Organization, error) {
	return s.GetOrganizationDBAL().GetAllOrganizations()
}

// GetOrganization returns a Organization by its ID
func (s *OrganizationService) GetOrganization(id string) (*asset.Organization, error) {
	return s.GetOrganizationDBAL().GetOrganization(id)
}
