package service

import (
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterMetric(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	mps := new(MockPermissionAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()

	provider.On("GetMetricDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetTimeService").Return(ts)
	provider.On("GetEventService").Return(es)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	service := NewMetricService(provider)

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	address := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	newPerms := &asset.NewPermissions{Public: true}

	metric := &asset.NewMetric{
		Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:           "Test metric",
		Address:        address,
		Description:    description,
		NewPermissions: newPerms,
	}

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_METRIC,
		AssetKey:  metric.Key,
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedMetric := &asset.Metric{
		Key:          "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:         "Test metric",
		Address:      address,
		Description:  description,
		Permissions:  perms,
		Owner:        "owner",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}

	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()
	dbal.On(
		"AddMetric",
		storedMetric,
	).Return(nil).Once()

	o, err := service.RegisterMetric(metric, "owner")

	assert.NoError(t, err, "Registration of valid metric should not fail")
	assert.NotNil(t, o, "Registration should return an Metric")
	assert.Equal(t, perms, o.Permissions, "Permissions should be set")
	assert.Equal(t, "owner", o.Owner, "Owner should be set")

	dbal.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestGetMetric(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetMetricDBAL").Return(dbal)
	service := NewMetricService(provider)

	metric := asset.Metric{
		Key:  "objKey",
		Name: "Test",
	}

	dbal.On("GetMetric", "objKey").Return(&metric, nil).Once()

	o, err := service.GetMetric("objKey")
	require.Nil(t, err)
	assert.Equal(t, o.Name, metric.Name)
}

func TestQueryMetrics(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetMetricDBAL").Return(dbal)
	service := NewMetricService(provider)

	obj1 := asset.Metric{
		Key:  "obj1",
		Name: "Test 1",
	}
	obj2 := asset.Metric{
		Key:  "obj2",
		Name: "Test 2",
	}

	pagination := common.NewPagination("", 12)

	dbal.On("QueryMetrics", pagination).Return([]*asset.Metric{&obj1, &obj2}, "nextPage", nil).Once()

	r, token, err := service.QueryMetrics(pagination)
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, obj1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestCanDownload(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetMetricDBAL").Return(dbal)
	service := NewMetricService(provider)

	perms := &asset.Permissions{
		Process: &asset.Permission{Public: true},
		Download: &asset.Permission{
			Public:        false,
			AuthorizedIds: []string{"org-2"},
		},
	}

	metric := &asset.Metric{
		Key:         "837B2E87-35CA-48F9-B83C-B40FB3FBA4E6",
		Name:        "Test",
		Permissions: perms,
	}

	dbal.On("GetMetric", "obj1").Return(metric, nil).Once()

	ok, err := service.CanDownload("obj1", "org-2")

	assert.Equal(t, ok, true)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
}
