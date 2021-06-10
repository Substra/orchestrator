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
)

func TestGetPlan(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	provider.On("GetComputePlanDBAL").Return(dbal)

	service := NewComputePlanService(provider)

	cp := &asset.ComputePlan{Key: "uuid", Owner: "org1", Tag: "test", TaskCount: 243, DoneCount: 223}

	dbal.On("GetComputePlan", "uuid").Once().Return(cp, nil)

	plan, err := service.GetPlan("uuid")
	assert.NoError(t, err)
	assert.Equal(t, cp, plan, "service should set task counters")

	dbal.AssertExpectations(t)
}

func TestRegisterPlan(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	es := new(MockEventService)
	provider := new(MockServiceProvider)

	provider.On("GetEventService").Return(es)
	provider.On("GetComputePlanDBAL").Return(dbal)

	service := NewComputePlanService(provider)

	newPlan := &asset.NewComputePlan{Key: "b9b3ecda-0a90-41da-a2e3-945eeafb06d8", Tag: "test"}

	expected := &asset.ComputePlan{
		Key:   "b9b3ecda-0a90-41da-a2e3-945eeafb06d8",
		Tag:   "test",
		Owner: "org1",
	}

	dbal.On("AddComputePlan", expected).Once().Return(nil)
	dbal.On("ComputePlanExists", "b9b3ecda-0a90-41da-a2e3-945eeafb06d8").Once().Return(false, nil)

	expectedEvent := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		AssetKey:  newPlan.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Metadata: map[string]string{
			"creator": "org1",
		},
	}
	es.On("RegisterEvents", []*asset.Event{expectedEvent}).Once().Return(nil)

	plan, err := service.RegisterPlan(newPlan, "org1")
	assert.NoError(t, err)
	assert.Equal(t, expected, plan)

	es.AssertExpectations(t)
	dbal.AssertExpectations(t)
}

func TestCancelPlan(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)

	service := NewComputePlanService(provider)

	plan := &asset.ComputePlan{
		Key:   "b9b3ecda-0a90-41da-a2e3-945eeafb06d8",
		Tag:   "test",
		Owner: "owner",
	}

	task1 := &asset.ComputeTask{Key: "uuid1"}
	task2 := &asset.ComputeTask{Key: "uuid2"}

	tasks := []*asset.ComputeTask{task1, task2}

	dbal.On("GetComputePlanTasks", "b9b3ecda-0a90-41da-a2e3-945eeafb06d8").Once().Return(tasks, nil)

	cts.On(
		"ApplyTaskAction",
		task1.Key,
		asset.ComputeTaskAction_TASK_ACTION_CANCELED,
		"compute plan b9b3ecda-0a90-41da-a2e3-945eeafb06d8 is cancelled",
		plan.Owner,
	).Once().Return(nil)
	cts.On(
		"ApplyTaskAction",
		task2.Key,
		asset.ComputeTaskAction_TASK_ACTION_CANCELED,
		"compute plan b9b3ecda-0a90-41da-a2e3-945eeafb06d8 is cancelled",
		plan.Owner,
	).Once().Return(&fsm.InvalidEventError{})

	err := service.cancelPlan(plan)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
}
