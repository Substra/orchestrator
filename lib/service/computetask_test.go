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
	"errors"
	"fmt"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orchestrationError "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
)

var newTrainTask = &asset.NewComputeTask{
	Key:            "867852b4-8419-4d52-8862-d5db823095be",
	Category:       asset.ComputeTaskCategory_TASK_TRAIN,
	AlgoKey:        "867852b4-8419-4d52-8862-d5db823095be",
	ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
	Data: &asset.NewComputeTask_Train{
		Train: &asset.NewTrainTaskData{
			DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
			DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
		},
	},
}

func TestGetTasks(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	pagination := common.NewPagination("", 2)
	filter := &asset.TaskQueryFilter{
		Status: asset.ComputeTaskStatus_STATUS_DOING,
	}

	returnedTasks := []*asset.ComputeTask{{}, {}}

	dbal.On("QueryComputeTasks", pagination, filter).Once().Return(returnedTasks, "", nil)

	tasks, _, err := service.GetTasks(pagination, filter)
	assert.NoError(t, err)

	assert.Len(t, tasks, 2)
}

func TestRegisterTaskConflict(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	dbal.On("ComputeTaskExists", newTrainTask.Key).Once().Return(true, nil)

	_, err := service.RegisterTask(newTrainTask, "test")
	assert.True(t, errors.Is(err, orchestrationError.ErrConflict))
}

func TestRegisterTrainTask(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	dispatcher := new(MockDispatcher)
	dms := new(MockDataManagerService)
	dss := new(MockDataSampleService)
	ps := new(MockPermissionService)
	as := new(MockAlgoService)

	provider.On("GetEventQueue").Return(dispatcher)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetDataSampleService").Return(dss)
	provider.On("GetPermissionService").Return(ps)
	provider.On("GetAlgoService").Return(as)

	service := NewComputeTaskService(provider)

	// Checking existing task
	dbal.On("ComputeTaskExists", newTrainTask.Key).Once().Return(false, nil)
	// checking parent compatibility (no parents here)
	dbal.On("GetComputeTasks", []string(nil)).Once().Return([]*asset.ComputeTask{}, nil)

	dataManagerKey := newTrainTask.Data.(*asset.NewComputeTask_Train).Train.DataManagerKey
	dataSampleKeys := newTrainTask.Data.(*asset.NewComputeTask_Train).Train.DataSampleKeys

	dataManager := &asset.DataManager{
		Key:   dataManagerKey,
		Owner: "dm-owner",
		Permissions: &asset.Permissions{
			Process:  &asset.Permission{Public: true},
			Download: &asset.Permission{Public: true},
		},
	}

	// Checking datamanager permissions
	dms.On("GetDataManager", dataManagerKey).Once().Return(dataManager, nil)
	ps.On("CanProcess", dataManager.Permissions, "testOwner").Once().Return(true)

	// Checking sample consistency
	dss.On("CheckSameManager", dataManagerKey, dataSampleKeys).Once().Return(nil)
	// Cannot train on test data
	dss.On("ContainsTestSample", dataSampleKeys).Once().Return(false, nil)

	algo := &asset.Algo{
		Category: asset.AlgoCategory_ALGO_SIMPLE,
		Permissions: &asset.Permissions{
			Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"testOwner"}},
			Download: &asset.Permission{Public: false, AuthorizedIds: []string{"testOwner"}},
		},
	}
	// check algo matching
	as.On("GetAlgo", newTrainTask.AlgoKey).Once().Return(algo, nil)
	ps.On("CanProcess", algo.Permissions, "testOwner").Once().Return(true)

	// create new permissions
	modelPerms := &asset.Permissions{
		Process:  &asset.Permission{AuthorizedIds: []string{"testOwner"}},
		Download: &asset.Permission{AuthorizedIds: []string{"testOwner"}},
	}
	ps.On("MergePermissions", algo.Permissions, dataManager.Permissions).Return(modelPerms)

	storedTask := &asset.ComputeTask{
		Key:            newTrainTask.Key,
		Category:       newTrainTask.Category,
		Algo:           algo,
		Owner:          "testOwner",
		ComputePlanKey: newTrainTask.ComputePlanKey,
		Metadata:       newTrainTask.Metadata,
		Rank:           newTrainTask.Rank,
		Status:         asset.ComputeTaskStatus_STATUS_TODO,
		ParentTaskKeys: newTrainTask.ParentTaskKeys,
		Worker:         dataManager.Owner,
		Data: &asset.ComputeTask_Train{
			Train: &asset.TrainTaskData{
				DataManagerKey:   dataManagerKey,
				DataSampleKeys:   dataSampleKeys,
				ModelKey:         "",
				ModelPermissions: modelPerms,
			},
		},
	}

	// finally store the created task
	dbal.On("AddComputeTask", storedTask).Once().Return(nil)

	expectedEvent := &event.Event{
		AssetKind: asset.ComputeTaskKind,
		AssetID:   newTrainTask.Key,
		EventKind: event.AssetCreated,
		Metadata: map[string]string{
			"status": storedTask.Status.String(),
		},
	}
	dispatcher.On("Enqueue", expectedEvent).Once().Return(nil)

	_, err := service.RegisterTask(newTrainTask, "testOwner")
	assert.NoError(t, err)
}

func TestRegisterFailedTask(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	dispatcher := new(MockDispatcher)
	dms := new(MockDataManagerService)
	dss := new(MockDataSampleService)
	ps := new(MockPermissionService)
	as := new(MockAlgoService)

	newTask := &asset.NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		ParentTaskKeys: []string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"},
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
				DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
			},
		},
	}

	provider.On("GetEventQueue").Return(dispatcher)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetDataSampleService").Return(dss)
	provider.On("GetPermissionService").Return(ps)
	provider.On("GetAlgoService").Return(as)

	service := NewComputeTaskService(provider)

	// Checking existing task
	dbal.On("ComputeTaskExists", newTrainTask.Key).Once().Return(false, nil)

	parentPerms := &asset.Permissions{Process: &asset.Permission{Public: true}}
	parentTask := &asset.ComputeTask{
		Status: asset.ComputeTaskStatus_STATUS_FAILED,
		Key:    "6c3878a8-8ca6-437e-83be-3a85b24b70d1",
		Data: &asset.ComputeTask_Train{
			Train: &asset.TrainTaskData{
				ModelPermissions: parentPerms,
			},
		},
	}
	// checking parent compatibility (a single failed parent)
	dbal.On("GetComputeTasks", []string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"}).Once().
		Return([]*asset.ComputeTask{parentTask}, nil)

	// Checking parent permissions
	ps.On("CanProcess", parentPerms, "testOwner").Once().Return(true)

	_, err := service.RegisterTask(newTask, "testOwner")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, orchestrationError.ErrIncompatibleTaskStatus))
}

func TestIsAlgoCompatible(t *testing.T) {
	cases := []struct {
		t asset.ComputeTaskCategory
		a asset.AlgoCategory
		o bool
	}{
		{t: asset.ComputeTaskCategory_TASK_AGGREGATE, a: asset.AlgoCategory_ALGO_AGGREGATE, o: true},
		{t: asset.ComputeTaskCategory_TASK_AGGREGATE, a: asset.AlgoCategory_ALGO_COMPOSITE, o: false},
		{t: asset.ComputeTaskCategory_TASK_COMPOSITE, a: asset.AlgoCategory_ALGO_COMPOSITE, o: true},
		{t: asset.ComputeTaskCategory_TASK_TRAIN, a: asset.AlgoCategory_ALGO_SIMPLE, o: true},
		{t: asset.ComputeTaskCategory_TASK_TRAIN, a: asset.AlgoCategory_ALGO_COMPOSITE, o: false},
		{t: asset.ComputeTaskCategory_TASK_TEST, a: asset.AlgoCategory_ALGO_COMPOSITE, o: true},
		{t: asset.ComputeTaskCategory_TASK_TEST, a: asset.AlgoCategory_ALGO_AGGREGATE, o: true},
		{t: asset.ComputeTaskCategory_TASK_TEST, a: asset.AlgoCategory_ALGO_SIMPLE, o: true},
	}

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("task %s and algo %s compat: %t", c.t.String(), c.a.String(), c.o),
			func(t *testing.T) {
				assert.Equal(t, c.o, isAlgoCompatible(c.t, c.a))
			},
		)
	}
}

func TestIsParentCompatible(t *testing.T) {
	cases := []struct {
		n string
		t asset.ComputeTaskCategory
		p []*asset.ComputeTask
		o bool
	}{
		{
			"Top train task",
			asset.ComputeTaskCategory_TASK_TRAIN,
			[]*asset.ComputeTask{},
			true, // Train can have no parent
		},
		{
			"Train task with test parent",
			asset.ComputeTaskCategory_TASK_TRAIN,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_TEST}},
			false, // Cannot train with a test parent
		},
		{
			"Train task with composite parent",
			asset.ComputeTaskCategory_TASK_TRAIN,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_COMPOSITE}},
			false, // Cannot train with a composite parent
		},
		{
			"Test task with composite parent",
			asset.ComputeTaskCategory_TASK_TEST,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_COMPOSITE}},
			true,
		},
		{
			"Test task with train parent",
			asset.ComputeTaskCategory_TASK_TEST,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_TRAIN}},
			true,
		},
		{
			"Test task with train and composite parent",
			asset.ComputeTaskCategory_TASK_TEST,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_TRAIN}, {Category: asset.ComputeTaskCategory_TASK_COMPOSITE}},
			false,
		},
		{
			"Aggregate task with train and composite parent",
			asset.ComputeTaskCategory_TASK_AGGREGATE,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_TRAIN}, {Category: asset.ComputeTaskCategory_TASK_COMPOSITE}},
			true,
		},
		{
			"Composite task with train and composite parent",
			asset.ComputeTaskCategory_TASK_COMPOSITE,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_TRAIN}, {Category: asset.ComputeTaskCategory_TASK_COMPOSITE}},
			false,
		},
		{
			"Composite task with train and composite parent",
			asset.ComputeTaskCategory_TASK_COMPOSITE,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_COMPOSITE}},
			true,
		},
		{
			"Composite task with aggregate and composite parent",
			asset.ComputeTaskCategory_TASK_COMPOSITE,
			[]*asset.ComputeTask{{Category: asset.ComputeTaskCategory_TASK_COMPOSITE}, {Category: asset.ComputeTaskCategory_TASK_AGGREGATE}},
			true,
		},
		{
			"Composite task without parents",
			asset.ComputeTaskCategory_TASK_COMPOSITE,
			[]*asset.ComputeTask{},
			true,
		},
	}

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("%s: %t", c.n, c.o),
			func(t *testing.T) {
				assert.Equal(t, c.o, isParentCompatible(c.t, c.p))
			},
		)
	}
}
