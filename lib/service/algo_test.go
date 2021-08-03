package service

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAlgo(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	provider.On("GetAlgoDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)

	service := NewAlgoService(provider)

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
		Category:       asset.AlgoCategory_ALGO_SIMPLE,
		Algorithm:      algorithm,
		Description:    description,
		NewPermissions: newPerms,
	}

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		AssetKey:  algo.Key,
	}
	es.On("RegisterEvents", e).Return(nil)

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedAlgo := &asset.Algo{
		Key:         "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:        "Test algo",
		Category:    asset.AlgoCategory_ALGO_SIMPLE,
		Algorithm:   algorithm,
		Description: description,
		Permissions: perms,
		Owner:       "owner",
	}

	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()
	dbal.On("AlgoExists", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(false, nil).Once()
	dbal.On(
		"AddAlgo",
		storedAlgo,
	).Return(nil).Once()

	o, err := service.RegisterAlgo(algo, "owner")

	assert.NoError(t, err, "Registration of valid algo should not fail")
	assert.NotNil(t, o, "Registration should return an Algo")
	assert.Equal(t, perms, o.Permissions, "Permissions should be set")
	assert.Equal(t, "owner", o.Owner, "Owner should be set")

	dbal.AssertExpectations(t)
}

func TestGetAlgo(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
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
	dbal := new(persistenceHelper.DBAL)
	provider := newMockedProvider()
	provider.On("GetAlgoDBAL").Return(dbal)
	service := NewAlgoService(provider)

	algo1 := asset.Algo{
		Key:      "algo1",
		Name:     "Test 1",
		Category: asset.AlgoCategory_ALGO_SIMPLE,
	}
	algo2 := asset.Algo{
		Key:      "algo2",
		Name:     "Test 2",
		Category: asset.AlgoCategory_ALGO_SIMPLE,
	}

	pagination := common.NewPagination("", 12)

	dbal.On("QueryAlgos", asset.AlgoCategory_ALGO_SIMPLE, pagination).Return([]*asset.Algo{&algo1, &algo2}, "nextPage", nil).Once()

	r, token, err := service.QueryAlgos(asset.AlgoCategory_ALGO_SIMPLE, pagination)
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, algo1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}
