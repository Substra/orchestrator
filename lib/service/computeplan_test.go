package service

import (
	"testing"

	"github.com/looplab/fsm"
	"github.com/owkin/orchestrator/lib/asset"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetPlan(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	provider := newMockedProvider()

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
	dbal := new(persistenceHelper.DBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

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
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)

	plan, err := service.RegisterPlan(newPlan, "org1")
	assert.NoError(t, err)
	assert.Equal(t, expected, plan)

	es.AssertExpectations(t)
	dbal.AssertExpectations(t)
}

func TestCancelPlan(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()

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

func TestComputePlanAllowIntermediaryModelDeletion(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	provider := newMockedProvider()

	provider.On("GetComputePlanDBAL").Return(dbal)

	service := NewComputePlanService(provider)

	cp := &asset.ComputePlan{
		Key:                      "uuid",
		DeleteIntermediaryModels: true,
	}

	dbal.On("GetRawComputePlan", "uuid").Once().Return(cp, nil)

	canDelete, err := service.canDeleteModels("uuid")
	assert.NoError(t, err)

	assert.True(t, canDelete)

	dbal.AssertExpectations(t)
}
