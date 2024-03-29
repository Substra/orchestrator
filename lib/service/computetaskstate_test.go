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

func TestGetInitialStatus(t *testing.T) {
	cases := map[string]struct {
		parents  []*asset.ComputeTask
		function *asset.Function
		outcome  asset.ComputeTaskStatus
	}{
		"no parents + function waiting": {
			parents:  []*asset.ComputeTask{},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_WAITING},
			outcome:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT,
		},
		"no parents + function building": {
			parents:  []*asset.ComputeTask{},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING},
			outcome:  asset.ComputeTaskStatus_STATUS_BUILDING,
		},
		"no parents + function ready": {
			parents:  []*asset.ComputeTask{},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_READY},
			outcome:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT,
		},
		"parent waiting + function waiting": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_WAITING},
			outcome:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT,
		},
		"parent waiting + function building": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING},
			outcome:  asset.ComputeTaskStatus_STATUS_BUILDING,
		},
		"parent waiting + function ready": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_READY},
			outcome:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS,
		},
		"parent ready + function waiting": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_WAITING},
			outcome:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT,
		},
		"parent ready + function building": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING},
			outcome:  asset.ComputeTaskStatus_STATUS_BUILDING,
		},
		"parent ready + function ready": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_READY},
			outcome:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT,
		},
		"parent failed": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_FAILED}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_READY},
			outcome:  asset.ComputeTaskStatus_STATUS_CANCELED,
		},
		"function failed": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_FAILED},
			outcome:  asset.ComputeTaskStatus_STATUS_FAILED,
		},
		"parent canceled": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_CANCELED}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_READY},
			outcome:  asset.ComputeTaskStatus_STATUS_CANCELED,
		},
		"function cancelled": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_CANCELED},
			outcome:  asset.ComputeTaskStatus_STATUS_CANCELED,
		},
		"parent canceled and failure": {
			parents:  []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_CANCELED}, {Status: asset.ComputeTaskStatus_STATUS_FAILED}},
			function: &asset.Function{Status: asset.FunctionStatus_FUNCTION_STATUS_READY},
			outcome:  asset.ComputeTaskStatus_STATUS_CANCELED,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, getInitialStatus(tc.parents, tc.function))
		})
	}
}

func TestOnStateChange(t *testing.T) {
	updater := new(mockTaskStateUpdater)
	updater.On("onStateChange", mock.Anything).Once()

	state := newState(updater, &asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT, Key: "uuid"})

	err := state.Event(context.Background(), string(transitionExecuting), &asset.ComputeTask{})

	assert.NoError(t, err)
	updater.AssertExpectations(t)
}

// Make sure fsm returns expected errors
func TestFailedStateChange(t *testing.T) {
	updater := new(mockTaskStateUpdater)

	state := newState(updater, &asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_EXECUTING, Key: "uuid"})

	err := state.Event(context.Background(), string(transitionExecuting), &asset.ComputeTask{})
	assert.IsType(t, fsm.InvalidEventError{}, err)

	state = newState(updater, &asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_DONE, Key: "uuid"})
	err = state.Event(context.Background(), string(transitionCanceled), &asset.ComputeTask{})
	assert.IsType(t, fsm.InvalidEventError{}, err)
	updater.AssertExpectations(t)
}

func TestDispatchOnTransition(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	service := NewComputeTaskService(provider)

	returnedTask := &asset.ComputeTask{
		Key:            "uuid",
		Status:         asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT,
		Worker:         "worker",
		ComputePlanKey: "uuidcp",
	}
	dbal.On("GetComputeTask", "uuid").Return(returnedTask, nil)

	expectedTask := &asset.ComputeTask{
		Key:            "uuid",
		Status:         asset.ComputeTaskStatus_STATUS_EXECUTING,
		Worker:         "worker",
		ComputePlanKey: "uuidcp",
	}
	dbal.On("UpdateComputeTaskStatus", expectedTask.Key, expectedTask.Status).Once().Return(nil)

	expectedEvent := &asset.Event{
		AssetKey:  "uuid",
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		Asset:     &asset.Event_ComputeTask{ComputeTask: expectedTask},
		Metadata: map[string]string{
			"reason": "User action",
		},
	}
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)

	err := service.ApplyTaskAction("uuid", asset.ComputeTaskAction_TASK_ACTION_EXECUTING, "", "worker")
	assert.NoError(t, err)

	es.AssertExpectations(t)
}

func TestUpdateTaskStateCanceled(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	// task is retrieved from persistence layer
	dbal.On("GetComputeTask", "uuid").Return(&asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS,
		Owner:  "owner",
	}, nil)
	// An update event should be enqueued
	es.On("RegisterEvents", mock.Anything).Return(nil)
	// Updated task should be saved
	updatedTask := &asset.ComputeTask{Key: "uuid", Status: asset.ComputeTaskStatus_STATUS_CANCELED, Owner: "owner"}
	dbal.On("UpdateComputeTaskStatus", updatedTask.Key, updatedTask.Status).Return(nil)

	service := NewComputeTaskService(provider)

	err := service.ApplyTaskAction("uuid", asset.ComputeTaskAction_TASK_ACTION_CANCELED, "", "owner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateTaskStateDone(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	dbal.On("GetComputeTask", "uuid").Return(&asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_EXECUTING,
		Owner:  "owner",
		Worker: "worker",
	}, nil)

	es.On("RegisterEvents", mock.Anything).Return(nil)

	updatedTask := &asset.ComputeTask{Key: "uuid", Status: asset.ComputeTaskStatus_STATUS_DONE, Owner: "owner", Worker: "worker"}

	dbal.On("UpdateComputeTaskStatus", updatedTask.Key, updatedTask.Status).Return(nil)

	dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{}, nil)

	service := NewComputeTaskService(provider)

	err := service.ApplyTaskAction("uuid", asset.ComputeTaskAction_TASK_ACTION_DONE, "", "worker")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestCascadeStatusDone(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	fs := new(MockFunctionAPI)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetFunctionService").Return(fs)

	task := &asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_EXECUTING,
		Owner:  "owner",
		Worker: "worker",
	}
	// Check for children to be updated
	dbal.On("GetComputeTaskParents", "child").Return([]*asset.ComputeTask{
		{Key: "uuid", Status: asset.ComputeTaskStatus_STATUS_DONE},
	}, nil)
	dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{
		{Key: "child", Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS},
	}, nil)

	// There should be two updates: 1 for the parent, 1 for the child
	es.On("RegisterEvents", mock.Anything).Times(2).Return(nil)

	// Updated task should be saved
	updatedParent := &asset.ComputeTask{Key: "uuid", Status: asset.ComputeTaskStatus_STATUS_DONE, Owner: "owner", Worker: "worker"}
	updatedChild := &asset.ComputeTask{Key: "child", Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT}
	dbal.On("UpdateComputeTaskStatus", updatedParent.Key, updatedParent.Status).Return(nil)
	dbal.On("UpdateComputeTaskStatus", updatedChild.Key, updatedChild.Status).Return(nil)

	service := NewComputeTaskService(provider)

	err := service.applyTaskTransition(task, transitionDone, "reason")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
	fs.AssertExpectations(t)
}

func TestUpdateAllowed(t *testing.T) {
	task := &asset.ComputeTask{
		Worker: "worker",
		Owner:  "owner",
	}
	cases := map[string]struct {
		requester string
		action    asset.ComputeTaskAction
		outcome   bool
	}{
		// BUILD_STARTED & BUILD_FINISHED should not be accessible from outside
		"owner build start": {
			requester: task.Owner,
			action:    asset.ComputeTaskAction_TASK_ACTION_BUILD_STARTED,
			outcome:   false,
		},
		"worker build start": {
			requester: task.Worker,
			action:    asset.ComputeTaskAction_TASK_ACTION_BUILD_STARTED,
			outcome:   false,
		},
		"owner build finish": {
			requester: task.Owner,
			action:    asset.ComputeTaskAction_TASK_ACTION_BUILD_FINISHED,
			outcome:   false,
		},
		"worker build finish": {
			requester: task.Worker,
			action:    asset.ComputeTaskAction_TASK_ACTION_BUILD_FINISHED,
			outcome:   false,
		},
		"owner executing": {
			requester: task.Owner,
			action:    asset.ComputeTaskAction_TASK_ACTION_EXECUTING,
			outcome:   false,
		},
		"worker executing": {
			requester: "worker",
			action:    asset.ComputeTaskAction_TASK_ACTION_EXECUTING,
			outcome:   true,
		},
		"owner cancel": {
			requester: "owner",
			action:    asset.ComputeTaskAction_TASK_ACTION_CANCELED,
			outcome:   true,
		},
		"worker cancel": {
			requester: "worker",
			action:    asset.ComputeTaskAction_TASK_ACTION_CANCELED,
			outcome:   true,
		},
		"owner failed": {
			requester: "owner",
			action:    asset.ComputeTaskAction_TASK_ACTION_FAILED,
			outcome:   true,
		},
		"worker failed": {
			requester: "worker",
			action:    asset.ComputeTaskAction_TASK_ACTION_FAILED,
			outcome:   true,
		},
		"owner done": {
			requester: "owner",
			action:    asset.ComputeTaskAction_TASK_ACTION_DONE,
			outcome:   false,
		},
		"worker done": {
			requester: "worker",
			action:    asset.ComputeTaskAction_TASK_ACTION_DONE,
			outcome:   true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, updateAllowed(task, tc.action, tc.requester))
		})
	}
}

func TestPropagateFunctionCancelation(t *testing.T) {
	cases := map[string]struct {
		taskOwner     string
		functionOwner string
	}{
		"same owner": {
			taskOwner:     "owner",
			functionOwner: "owner",
		},
		"different owners": {
			taskOwner:     "taskOwner",
			functionOwner: "functionOwner",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbal := new(persistence.MockDBAL)
			es := new(MockEventAPI)
			cps := new(MockComputePlanAPI)
			provider := newMockedProvider()
			service := NewComputeTaskService(provider)

			provider.On("GetComputeTaskDBAL").Return(dbal)
			provider.On("GetEventService").Return(es)
			provider.On("GetComputePlanService").Return(cps)

			functionKey := "uuid_f"
			task := &asset.ComputeTask{Key: "uuid_t", Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT, Owner: tc.taskOwner, Worker: "worker"}

			cps.On("failPlan", mock.Anything).Return(nil)
			dbal.On("GetFunctionFromTasksWithStatus", functionKey, []asset.ComputeTaskStatus{
				asset.ComputeTaskStatus_STATUS_BUILDING,
			}).Return([]*asset.ComputeTask{task}, nil)
			dbal.On("UpdateComputeTaskStatus", task.Key, asset.ComputeTaskStatus_STATUS_FAILED).Return(nil)
			es.On("RegisterEvents", mock.Anything).Return(nil)

			cps.On("failPlan", mock.Anything).Return(nil)
			dbal.On("GetFunctionFromTasksWithStatus", functionKey, []asset.ComputeTaskStatus{
				asset.ComputeTaskStatus_STATUS_BUILDING,
			}).Return([]*asset.ComputeTask{task}, nil)
			dbal.On("UpdateComputeTaskStatus", task.Key, asset.ComputeTaskStatus_STATUS_FAILED).Return(nil)
			es.On("RegisterEvents", mock.Anything).Return(nil)

			err := service.PropagateActionFromFunction(functionKey, asset.ComputeTaskAction_TASK_ACTION_FAILED, "Building failed", tc.functionOwner)

			assert.NoError(t, err)

			cps.AssertExpectations(t)
			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
			provider.AssertExpectations(t)
		})
	}
}

func TestCheckParentDone(t *testing.T) {
	cases := map[string]struct {
		parents []*asset.ComputeTask
		outcome bool
	}{
		"no parents": {
			parents: []*asset.ComputeTask{},
			outcome: true,
		},
		"2 parents done": {
			parents: []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			outcome: true,
		},
		"1 parent done + 1 parent waiting": {
			parents: []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS}},
			outcome: false,
		},
	}

	childKey := "childKey"

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			provider := newMockedProvider()
			service := NewComputeTaskService(provider)
			dbal := new(persistence.MockDBAL)

			provider.On("GetComputeTaskDBAL").Return(dbal).Once()
			dbal.On("GetComputeTaskParents", childKey).Return(tc.parents, nil).Once()

			done, err := service.checkParentTasksDone(childKey)

			provider.AssertExpectations(t)
			dbal.AssertExpectations(t)
			assert.NoError(t, err)
			assert.Equal(t, tc.outcome, done)
		})
	}
}
