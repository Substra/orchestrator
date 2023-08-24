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

	state := newFunctionState(updater, &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_CREATED, Key: "uuid"})

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
		Key:            "uuid",
		Status:         asset.FunctionStatus_FUNCTION_STATUS_CREATED,
		Owner:          "owner",
	}
	dbal.On("GetFunction", "uuid").Return(returnedFunction, nil)

	expectedFunction := &asset.Function{
		Key:            "uuid",
		Status:         asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
		Owner:          "owner",
	}
	dbal.On("UpdateFunctionStatus", expectedFunction.Key, expectedFunction.Status).Once().Return(nil)

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

	es.AssertExpectations(t)
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
		Status: asset.FunctionStatus_FUNCTION_STATUS_CREATED,
		Owner:  "owner",
	}, nil)
	// An update event should be enqueued
	es.On("RegisterEvents", mock.Anything).Return(nil)
	// Updated function should be saved
	updatedFunction := &asset.Function{Key: "uuid", Status: asset.FunctionStatus_FUNCTION_STATUS_CANCELED, Owner: "owner"}
	dbal.On("UpdateFunctionStatus", updatedFunction.Key, updatedFunction.Status).Return(nil)

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
	
	taskKey := "uuid_ct"
	ct.On("QueryTasks", mock.Anything, &asset.TaskQueryFilter{Status: asset.ComputeTaskStatus_STATUS_DOING, FunctionKey:functionKey}).Return([]*asset.ComputeTask{
		{
			Key: taskKey,
			Owner: "owner",
			FunctionKey: functionKey,
			Status: asset.ComputeTaskStatus_STATUS_DOING,
		},
	}, "", nil)
	ct.On("ApplyTaskAction", taskKey, asset.ComputeTaskAction_TASK_ACTION_FAILED, mock.Anything, "owner").Return(nil)
	es.On("RegisterEvents", mock.Anything).Return(nil)

	updatedFunction := &asset.Function{Key: functionKey, Status: asset.FunctionStatus_FUNCTION_STATUS_FAILED, Owner: "owner"}

	dbal.On("UpdateFunctionStatus", updatedFunction.Key, updatedFunction.Status).Return(nil)


	service := NewFunctionService(provider)

	err := service.ApplyFunctionAction("uuid", asset.FunctionAction_FUNCTION_ACTION_FAILED, "", "owner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}
