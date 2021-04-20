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
	"github.com/owkin/orchestrator/lib/event"
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

func (m *mockStateUpdater) cascadeCancel(e *fsm.Event) {
	m.Called(e)
}

func (m *mockStateUpdater) updateChildrenStatus(e *fsm.Event) {
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

	err := state.Event(asset.ComputeTaskAction_TASK_ACTION_DOING.String(), &asset.ComputeTask{})

	assert.NoError(t, err)
}

func TestDispatchOnTransition(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)

	service := NewComputeTaskService(provider)

	returnedTask := &asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_TODO,
	}
	dbal.On("GetComputeTask", "uuid").Once().Return(returnedTask, nil)

	expectedTask := &asset.ComputeTask{
		Key:    "uuid",
		Status: asset.ComputeTaskStatus_STATUS_DOING,
	}
	dbal.On("UpdateComputeTask", expectedTask).Once().Return(nil)

	expectedEvent := &event.Event{
		AssetID:   "uuid",
		AssetKind: asset.ComputeTaskKind,
		EventKind: event.AssetUpdated,
		Metadata: map[string]string{
			"status": expectedTask.Status.String(),
			"reason": "User action",
		},
	}

	dispatcher.On("Enqueue", expectedEvent).Once().Return(nil)

	err := service.ApplyTaskAction("uuid", asset.ComputeTaskAction_TASK_ACTION_DOING, "")
	assert.NoError(t, err)
}

func TestUpdateTaskStateCanceled(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)

	// task is retrieved from persistence layer
	dbal.On("GetComputeTask", "uuid").Return(&asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_WAITING}, nil)
	// An update event should be enqueued
	dispatcher.On("Enqueue", mock.Anything).Return(nil)
	// Updated task should be saved
	updatedTask := &asset.ComputeTask{Status: asset.ComputeTaskStatus_STATUS_CANCELED}
	dbal.On("UpdateComputeTask", updatedTask).Return(nil)

	service := NewComputeTaskService(provider)

	err := service.ApplyTaskAction("uuid", asset.ComputeTaskAction_TASK_ACTION_CANCELED, "")
	assert.NoError(t, err)
}
