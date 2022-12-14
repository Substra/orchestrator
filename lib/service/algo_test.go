package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterAlgo(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	mps := new(MockPermissionAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()

	provider.On("GetAlgoDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)

	service := NewAlgoService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	algorithm := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	newPerms := &asset.NewPermissions{Public: true}

	algo := &asset.NewAlgo{
		Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:           "Test algo",
		Algorithm:      algorithm,
		Description:    description,
		NewPermissions: newPerms,
	}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}
	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()

	dbal.On("AlgoExists", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(false, nil).Once()

	storedAlgo := &asset.Algo{
		Key:          "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:         "Test algo",
		Algorithm:    algorithm,
		Description:  description,
		Permissions:  perms,
		Owner:        "owner",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddAlgo", storedAlgo).Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		AssetKey:  algo.Key,
		Asset:     &asset.Event_Algo{Algo: storedAlgo},
	}
	es.On("RegisterEvents", e).Return(nil)

	o, err := service.RegisterAlgo(algo, "owner")

	assert.NoError(t, err, "Registration of valid algo should not fail")
	assert.NotNil(t, o, "Registration should return an Algo")
	assert.Equal(t, perms, o.Permissions, "Permissions should be set")
	assert.Equal(t, "owner", o.Owner, "Owner should be set")

	dbal.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestGetAlgo(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetAlgoDBAL").Return(dbal)
	service := NewAlgoService(provider)

	algo := asset.Algo{
		Key:  "algoKey",
		Name: "Test",
	}

	dbal.On("GetAlgo", "algoKey").Return(&algo, nil).Once()

	o, err := service.GetAlgo("algoKey")
	require.Nil(t, err)
	assert.Equal(t, o.Name, algo.Name)
}

func TestQueryAlgos(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetAlgoDBAL").Return(dbal)
	service := NewAlgoService(provider)

	computePlanKey := uuid.NewString()

	algo1 := asset.Algo{
		Key:  "algo1",
		Name: "Test 1",
	}
	algo2 := asset.Algo{
		Key:  "algo2",
		Name: "Test 2",
	}

	pagination := common.NewPagination("", 12)

	dbal.On("QueryAlgos", pagination, &asset.AlgoQueryFilter{ComputePlanKey: computePlanKey}).
		Return([]*asset.Algo{&algo1, &algo2}, "nextPage", nil).Once()

	r, token, err := service.QueryAlgos(pagination, &asset.AlgoQueryFilter{ComputePlanKey: computePlanKey})
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, algo1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestCanDownload(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetAlgoDBAL").Return(dbal)
	service := NewAlgoService(provider)

	perms := &asset.Permissions{
		Process: &asset.Permission{Public: true},
		Download: &asset.Permission{
			Public:        false,
			AuthorizedIds: []string{"org-2"},
		},
	}

	algo := &asset.Algo{
		Key:         "837B2E87-35CA-48F9-B83C-B40FB3FBA4E6",
		Name:        "Test",
		Permissions: perms,
	}

	dbal.On("GetAlgo", "obj1").Return(algo, nil).Once()

	ok, err := service.CanDownload("obj1", "org-2")

	assert.Equal(t, ok, true)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
}

func TestUpdateSingleExistingAlgo(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetAlgoDBAL").Return(dbal)
	service := NewAlgoService(provider)

	existingAlgo := &asset.Algo{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "algo name",
		Owner: "owner",
	}

	updateAlgoParam := &asset.UpdateAlgoParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated algo name",
	}

	storedAlgo := &asset.Algo{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "Updated algo name",
		Owner: "owner",
	}

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		AssetKey:  storedAlgo.Key,
		Asset:     &asset.Event_Algo{Algo: storedAlgo},
	}

	cases := map[string]struct {
		requester string
		valid     bool
	}{
		"update successful": {
			requester: "owner",
			valid:     true,
		},
		"update rejected: requester is not owner": {
			requester: "user",
			valid:     false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbal.On("GetAlgo", existingAlgo.GetKey()).Return(existingAlgo, nil).Once()

			if tc.valid {
				dbal.On("UpdateAlgo", storedAlgo).Return(nil).Once()
				es.On("RegisterEvents", e).Once().Return(nil)
			}

			err := service.UpdateAlgo(updateAlgoParam, tc.requester)

			if tc.valid {
				assert.NoError(t, err, "Update of algo should not fail")
			} else {
				assert.Error(t, err, "Update of algo should fail")
			}

			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
		})
	}

}
