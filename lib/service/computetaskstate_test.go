// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"testing"

	"github.com/looplab/fsm"
	"github.com/owkin/orchestrator/lib/asset"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStateUpdater struct {
	mock.Mock
}

func (m *mockStateUpdater) onStateChange(e *fsm.Event) {
	m.Called(e)
}

func (m *mockStateUpdater) onCancel(e *fsm.Event) {
	m.Called(e)
}

func (m *mockStateUpdater) onDone(e *fsm.Event) {
	m.Called(e)
}

func TestGetInitialStatusFromParents(t *testing.T) {
	cases := map[string]struct {
		parents []*asset.ComputeTask
		outcome asset.ComputeTaskStatus
	}{
		"no parents": {
			parents: []*asset.ComputeTask{},
			outcome: asset.ComputeTaskStatus_STATUS_TODO,
		},
		"waiting": {
			parents: []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_WAITING}},
			outcome: asset.ComputeTaskStatus_STATUS_WAITING,
		},
		"ready": {
			parents: []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_DONE}},
			outcome: asset.ComputeTaskStatus_STATUS_TODO,
		},
		"failed": {
			parents: []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_FAILED}},
			outcome: asset.ComputeTaskStatus_STATUS_CANCELED,
		},
		"canceled": {
			parents: []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_CANCELED}},
			outcome: asset.ComputeTaskStatus_STATUS_CANCELED,
		},
		"canceled and failure": {
			parents: []*asset.ComputeTask{{Status: asset.ComputeTaskStatus_STATUS_DONE}, {Status: asset.ComputeTaskStatus_STATUS_CANCELED}, {Status: asset.ComputeTaskStatus_STATUS_FAILED}},
			outcome: asset.ComputeTaskStatus_STATUS_CANCELED,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, getInitialStatusFromParents(tc.parents))
		})
	}
}

func TestOnStateChange(t *testing.T) {
	updater := new(mockStateUpdater)
	updater.On("onStateChange", mock.Anything).Once()

	state := newState(updater, &asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_TODO, Key: "uuid"})

	err := state.Event(string(transitionDoing), &asset.ComputeTask{})

	assert.NoError(t, err)
	updater.AssertExpectations(t)
}

// Make sure fsm returns expected errors
func TestFailedStateChange(t *testing.T) {
	updater := new(mockStateUpdater)

	state := newState(updater, &asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_DOING, Key: "uuid"})

	err := state.Event(string(transitionDoing), &asset.ComputeTask{})
	assert.IsType(t, fsm.InvalidEventError{}, err)

	state = newState(updater, &asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_DONE, Key: "uuid"})
	err = state.Event(string(transitionCanceled), &asset.ComputeTask{})
	assert.IsType(t, fsm.InvalidEventError{}, err)
	updater.AssertExpectations(t)
}

func TestDispatchOnTransition(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	es := new(MockEventService)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	service := NewComputeTaskService(provider)

	returnedTask := &asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_TODO,
		Worker: "worker",
	}
	dbal.On("GetComputeTask", "uuid").Return(returnedTask, nil)

	expectedTask := &asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_DOING,
		Worker: "worker",
	}
	dbal.On("UpdateComputeTask", expectedTask).Once().Return(nil)

	expectedEvent := &asset.Event{
		AssetKey:  "uuid",
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		Metadata: map[string]string{
			"status": expectedTask.Status.String(),
			"reason": "User action",
		},
	}
	es.On("RegisterEvent", expectedEvent).Once().Return(nil)

	err := service.ApplyTaskAction("uuid", asset.ComputeTaskAction_TASK_ACTION_DOING, "", "worker")
	assert.NoError(t, err)

	es.AssertExpectations(t)
}

func TestUpdateTaskStateCanceled(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	es := new(MockEventService)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	// task is retrieved from persistence layer
	dbal.On("GetComputeTask", "uuid").Return(&asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_WAITING,
		Owner:  "owner",
	}, nil)
	// Check for children to be updated
	dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{}, nil)
	// An update event should be enqueued
	es.On("RegisterEvent", mock.Anything).Return(nil)
	// Updated task should be saved
	updatedTask := &asset.ComputeTask{Key: "uuid", Status: asset.ComputeTaskStatus_STATUS_CANCELED, Owner: "owner"}
	dbal.On("UpdateComputeTask", updatedTask).Return(nil)

	service := NewComputeTaskService(provider)

	err := service.ApplyTaskAction("uuid", asset.ComputeTaskAction_TASK_ACTION_CANCELED, "", "owner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestCascadeStatusDone(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	es := new(MockEventService)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	task := &asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_DOING,
		Owner:  "owner",
		Worker: "worker",
	}
	// Check for children to be updated
	dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{
		{Key: "child", Status: asset.ComputeTaskStatus_STATUS_WAITING},
	}, nil)

	// There should be two updates: 1 for the parent, 1 for the child
	es.On("RegisterEvent", mock.Anything).Times(2).Return(nil)
	// Updated task should be saved
	updatedParent := &asset.ComputeTask{Key: "uuid", Status: asset.ComputeTaskStatus_STATUS_DONE, Owner: "owner", Worker: "worker"}
	updatedChild := &asset.ComputeTask{Key: "child", Status: asset.ComputeTaskStatus_STATUS_TODO}
	dbal.On("UpdateComputeTask", updatedParent).Return(nil)
	dbal.On("UpdateComputeTask", updatedChild).Return(nil)

	service := NewComputeTaskService(provider)

	err := service.applyTaskAction(task, transitionDone, "reason")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
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
		"owner cancel": {
			requester: "owner",
			action:    asset.ComputeTaskAction_TASK_ACTION_CANCELED,
			outcome:   true,
		},
		"worker cancel": {
			requester: "worker",
			action:    asset.ComputeTaskAction_TASK_ACTION_CANCELED,
			outcome:   false,
		},
		"worker fail": {
			requester: "worker",
			action:    asset.ComputeTaskAction_TASK_ACTION_FAILED,
			outcome:   true,
		},
		"worker doing": {
			requester: "worker",
			action:    asset.ComputeTaskAction_TASK_ACTION_DOING,
			outcome:   true,
		},
		"owner doing": {
			requester: "owner",
			action:    asset.ComputeTaskAction_TASK_ACTION_DOING,
			outcome:   false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, updateAllowed(task, tc.action, tc.requester))
		})
	}
}
