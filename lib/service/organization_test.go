package service

import (
	"errors"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterOrganization(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)

	provider.On("GetOrganizationDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	expected := &asset.Organization{
		Id:           "uuid1",
		Address:      "org-1.com",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}

	dbal.On("OrganizationExists", "uuid1").Return(false, nil).Once()
	dbal.On("AddOrganization", expected).Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_ORGANIZATION,
		AssetKey:  "uuid1",
		Asset:     &asset.Event_Organization{Organization: expected},
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	service := NewOrganizationService(provider)

	newOrganization := &asset.RegisterOrganizationParam{
		Address: "org-1.com",
	}

	organization, err := service.RegisterOrganization("uuid1", newOrganization)
	assert.NoError(t, err, "Organization registration should not fail")
	assert.Equal(t, expected, organization, "Registration should return a organization")

	dbal.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterExistingOrganization(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetOrganizationDBAL").Return(dbal)

	dbal.On("OrganizationExists", "uuid1").Return(true, nil).Once()

	service := NewOrganizationService(provider)

	newOrganization := &asset.RegisterOrganizationParam{
		Address: "org-1.com",
	}

	_, err := service.RegisterOrganization("uuid1", newOrganization)
	assert.Error(t, err, "Registration should fail for existing organization")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrConflict, orcError.Kind)
}
