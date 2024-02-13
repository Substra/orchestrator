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
	cases := map[string]struct {
		functionStatusBefore         asset.FunctionStatus
		functionAction               asset.FunctionAction
		expectedFunctionStatusAfter  asset.FunctionStatus
		expectedPropagatedTaskAction asset.ComputeTaskAction
	}{
		"build started": {
			functionStatusBefore:         asset.FunctionStatus_FUNCTION_STATUS_WAITING,
			functionAction:               asset.FunctionAction_FUNCTION_ACTION_BUILDING,
			expectedFunctionStatusAfter:  asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
			expectedPropagatedTaskAction: asset.ComputeTaskAction_TASK_ACTION_BUILD_STARTED,
		},
		"failed": {
			functionStatusBefore:         asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
			functionAction:               asset.FunctionAction_FUNCTION_ACTION_FAILED,
			expectedFunctionStatusAfter:  asset.FunctionStatus_FUNCTION_STATUS_FAILED,
			expectedPropagatedTaskAction: asset.ComputeTaskAction_TASK_ACTION_FAILED,
		},
		"canceled": {
			functionStatusBefore:         asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
			functionAction:               asset.FunctionAction_FUNCTION_ACTION_CANCELED,
			expectedFunctionStatusAfter:  asset.FunctionStatus_FUNCTION_STATUS_CANCELED,
			expectedPropagatedTaskAction: asset.ComputeTaskAction_TASK_ACTION_CANCELED,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbal := new(persistence.MockDBAL)
			ct := new(MockComputeTaskAPI)
			es := new(MockEventAPI)
			provider := newMockedProvider()

			provider.On("GetFunctionDBAL").Return(dbal)
			provider.On("GetComputeTaskService").Return(ct)
			provider.On("GetEventService").Return(es)

			service := NewFunctionService(provider)

			returnedFunction := &asset.Function{
				Key:    "uuid",
				Status: tc.functionStatusBefore,
				Owner:  "owner",
			}
			dbal.On("GetFunction", returnedFunction.Key).Return(returnedFunction, nil)

			expectedFunction := &asset.Function{
				Key:    returnedFunction.Key,
				Status: tc.expectedFunctionStatusAfter,
				Owner:  returnedFunction.Owner,
			}
			dbal.On("UpdateFunction", expectedFunction).Once().Return(nil)

			expectedEvent := &asset.Event{
				AssetKey:  returnedFunction.Key,
				AssetKind: asset.AssetKind_ASSET_FUNCTION,
				EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
				Asset:     &asset.Event_Function{Function: expectedFunction},
				Metadata: map[string]string{
					"reason": "User action",
				},
			}
			es.On("RegisterEvents", expectedEvent).Once().Return(nil)
			ct.On("PropagateActionFromFunction", "uuid", tc.expectedPropagatedTaskAction, "User action", expectedFunction.Owner).Return(nil).Once()

			err := service.ApplyFunctionAction("uuid", tc.functionAction, "", expectedFunction.Owner)
			assert.NoError(t, err)

			ct.AssertExpectations(t)
			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
			provider.AssertExpectations(t)
		})
	}
}

func TestUpdateFunctionStateReady(t *testing.T) {
	cases := map[string]struct {
		parents            []*asset.ComputeTask
		expectedTaskStatus asset.ComputeTaskStatus
	}{
		"no parents": {
			parents:            []*asset.ComputeTask{},
			expectedTaskStatus: asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT,
		},
		"2 parents done": {
			parents:            []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			expectedTaskStatus: asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT,
		},
		"1 parent done + 1 parent waiting": {
			parents:            []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS}},
			expectedTaskStatus: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS,
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
				Status:         asset.ComputeTaskStatus_STATUS_BUILDING,
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
			ctdbal.On("GetFunctionFromTasksWithStatus", functionKey, []asset.ComputeTaskStatus{asset.ComputeTaskStatus_STATUS_BUILDING}).Return([]*asset.ComputeTask{task}, nil)
			ctdbal.On("GetComputeTaskParents", task.Key).Return(tc.parents, nil).Once()
			ctdbal.On("UpdateComputeTaskStatus", task.Key, tc.expectedTaskStatus).Return(nil).Once()

			err := service.ApplyFunctionAction("uuid", asset.FunctionAction_FUNCTION_ACTION_READY, "", "owner")
			assert.NoError(t, err)

			provider.AssertExpectations(t)
			ctdbal.AssertExpectations(t)
			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
		})
	}
}
