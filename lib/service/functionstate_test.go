package service

import (
	"context"
	"testing"

	"github.com/looplab/fsm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/persistence"
)

func TestOnFunctionStateChange(t *testing.T) {
	updater := new(mockFunctionStateUpdater)
	updater.On("onStateChange", mock.Anything).Once()

	state := newFunctionState(updater, &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_WAITING, Key: "uuid"})

	err := state.Event(context.Background(), string(transitionFunctionBuilding), &asset.Function{})

	assert.NoError(t, err)
	updater.AssertExpectations(t)
}

// Make sure fsm returns expected errors
func TestFailedFunctionStateChange(t *testing.T) {
	updater := new(mockFunctionStateUpdater)

	state := newFunctionState(updater, &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING, Key: "uuid"})

	err := state.Event(context.Background(), string(transitionFunctionBuilding), &asset.Function{})
	assert.IsType(t, fsm.InvalidEventError{}, err)

	state = newFunctionState(updater, &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_READY, Key: "uuid"})
	err = state.Event(context.Background(), string(transitionFunctionCanceled), &asset.Function{})
	assert.IsType(t, fsm.InvalidEventError{}, err)
	updater.AssertExpectations(t)
}

func TestDispatchOnFunctionTransition(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	provider.On("GetFunctionDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	service := NewFunctionService(provider)

	returnedFunction := &asset.Function{
		Key:    "uuid",
		Status: asset.FunctionStatus_FUNCTION_STATUS_WAITING,
		Owner:  "owner",
	}
	dbal.On("GetFunction", "uuid").Return(returnedFunction, nil)

	expectedFunction := &asset.Function{
		Key:    "uuid",
		Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
		Owner:  "owner",
	}
	dbal.On("UpdateFunction", expectedFunction).Once().Return(nil)

	expectedEvent := &asset.Event{
		AssetKey:  "uuid",
		AssetKind: asset.AssetKind_ASSET_FUNCTION,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		Asset:     &asset.Event_Function{Function: expectedFunction},
		Metadata: map[string]string{
			"reason": "User action",
		},
	}
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)

	err := service.ApplyFunctionAction("uuid", asset.FunctionAction_FUNCTION_ACTION_BUILDING, "", "owner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
	provider.AssertExpectations(t)
}

// Testing that failing a Function propagate to tasks using this function
func TestUpdateFunctionStateCanceled(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	provider.On("GetFunctionDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	// function is retrieved from persistence layer
	dbal.On("GetFunction", "uuid").Return(&asset.Function{
		Key:    "uuid",
		Status: asset.FunctionStatus_FUNCTION_STATUS_WAITING,
		Owner:  "owner",
	}, nil)
	// An update event should be enqueued
	es.On("RegisterEvents", mock.Anything).Return(nil)
	// Updated function should be saved
	updatedFunction := &asset.Function{Key: "uuid", Status: asset.FunctionStatus_FUNCTION_STATUS_CANCELED, Owner: "owner"}
	dbal.On("UpdateFunction", updatedFunction).Return(nil)

	service := NewFunctionService(provider)

	err := service.ApplyFunctionAction("uuid", asset.FunctionAction_FUNCTION_ACTION_CANCELED, "", "owner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateFunctionStateFailed(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	ct := new(MockComputeTaskAPI)
	provider := newMockedProvider()

	provider.On("GetFunctionDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(ct)
	provider.On("GetEventService").Return(es)

	functionKey := "uuid"

	dbal.On("GetFunction", "uuid").Return(&asset.Function{
		Key:    functionKey,
		Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
		Owner:  "owner",
	}, nil)

	ct.On("propagateFunctionCancelation", functionKey, "owner").Return(nil)
	es.On("RegisterEvents", mock.Anything).Return(nil)

	updatedFunction := &asset.Function{Key: functionKey, Status: asset.FunctionStatus_FUNCTION_STATUS_FAILED, Owner: "owner"}

	dbal.On("UpdateFunction", updatedFunction).Return(nil)

	service := NewFunctionService(provider)

	err := service.ApplyFunctionAction("uuid", asset.FunctionAction_FUNCTION_ACTION_FAILED, "", "owner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateFunctionStateReady(t *testing.T) {
	cases := map[string]struct {
		parents    []*asset.ComputeTask
		becomeTodo bool
	}{
		"no parents": {
			parents:    []*asset.ComputeTask{},
			becomeTodo: true,
		},
		"2 parents done": {
			parents:    []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			becomeTodo: true,
		},
		"1 parent done + 1 parent waiting": {
			parents:    []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_WAITING}},
			becomeTodo: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbal := new(persistence.MockDBAL)
			ctdbal := new(persistence.MockDBAL)
			es := new(MockEventAPI)
			provider := newMockedProvider()

			service := NewFunctionService(provider)

			provider.On("GetFunctionDBAL").Return(dbal)
			provider.On("GetComputeTaskService").Return(NewComputeTaskService(provider))
			provider.On("GetEventService").Return(es)
			provider.On("GetComputeTaskDBAL").Return(ctdbal)

			functionKey := "uuid"
			task := &asset.ComputeTask{
				Key:            "taskuuid",
				Status:         asset.ComputeTaskStatus_STATUS_WAITING,
				Worker:         "worker",
				ComputePlanKey: "uuidcp",
			}

			dbal.On("GetFunction", "uuid").Return(&asset.Function{
				Key:    functionKey,
				Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
				Owner:  "owner",
			}, nil)

			es.On("RegisterEvents", mock.Anything).Return(nil)

			updatedFunction := &asset.Function{Key: functionKey, Status: asset.FunctionStatus_FUNCTION_STATUS_READY, Owner: "owner"}

			dbal.On("UpdateFunction", updatedFunction).Return(nil)
			ctdbal.On("GetFunctionFromTasksWithStatus", functionKey, []asset.ComputeTaskStatus{asset.ComputeTaskStatus_STATUS_WAITING}).Return([]*asset.ComputeTask{task}, nil)
			ctdbal.On("GetComputeTaskParents", task.Key).Return(tc.parents, nil).Once()

			if tc.becomeTodo {
				ctdbal.On("UpdateComputeTaskStatus", task.Key, asset.ComputeTaskStatus_STATUS_TODO).Return(nil).Once()
			}

			err := service.ApplyFunctionAction("uuid", asset.FunctionAction_FUNCTION_ACTION_READY, "", "owner")
			assert.NoError(t, err)

			provider.AssertExpectations(t)
			ctdbal.AssertExpectations(t)
			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
		})
	}
}
