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

func TestRegisterNode(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)

	provider.On("GetNodeDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	expected := &asset.Node{
		Id:           "uuid1",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}

	dbal.On("NodeExists", "uuid1").Return(false, nil).Once()
	dbal.On("AddNode", expected).Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_NODE,
		AssetKey:  "uuid1",
		Asset:     &asset.Event_Node{Node: expected},
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	service := NewNodeService(provider)

	node, err := service.RegisterNode("uuid1")
	assert.NoError(t, err, "Node registration should not fail")
	assert.Equal(t, expected, node, "Registration should return a node")

	dbal.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterExistingNode(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetNodeDBAL").Return(dbal)

	dbal.On("NodeExists", "uuid1").Return(true, nil).Once()

	service := NewNodeService(provider)

	_, err := service.RegisterNode("uuid1")
	assert.Error(t, err, "Registration should fail for existing node")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrConflict, orcError.Kind)
}
