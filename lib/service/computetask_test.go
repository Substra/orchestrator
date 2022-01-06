package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func TestGetTask(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	task := &asset.ComputeTask{
		Key: "uuid",
	}

	dbal.On("GetComputeTask", "uuid").Once().Return(task, nil)

	ret, err := service.GetTask("uuid")
	assert.NoError(t, err)
	assert.Equal(t, task, ret)

	provider.AssertExpectations(t)
	dbal.AssertExpectations(t)
}

func TestQueryTasks(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	pagination := common.NewPagination("", 2)
	filter := &asset.TaskQueryFilter{
		Status: asset.ComputeTaskStatus_STATUS_DOING,
	}

	returnedTasks := []*asset.ComputeTask{{}, {}}

	dbal.On("QueryComputeTasks", pagination, filter).Once().Return(returnedTasks, "", nil)

	tasks, _, err := service.QueryTasks(pagination, filter)
	assert.NoError(t, err)

	assert.Len(t, tasks, 2)
}

func TestRegisterTaskConflict(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	dbal.On("GetExistingComputeTaskKeys", []string{}).Once().Return([]string{}, nil)
	dbal.On("ComputeTaskExists", newTrainTask.Key).Once().Return(true, nil)

	err := service.RegisterTasks([]*asset.NewComputeTask{newTrainTask}, "test")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrConflict, orcError.Kind)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterTrainTask(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	dms := new(MockDataManagerAPI)
	dss := new(MockDataSampleAPI)
	ps := new(MockPermissionAPI)
	as := new(MockAlgoAPI)
	ts := new(MockTimeAPI)

	provider.On("GetEventService").Return(es)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetDataSampleService").Return(dss)
	provider.On("GetPermissionService").Return(ps)
	provider.On("GetAlgoService").Return(as)
	provider.On("GetTimeService").Return(ts)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	service := NewComputeTaskService(provider)

	// Checking existing task
	dbal.On("ComputeTaskExists", newTrainTask.Key).Once().Return(false, nil)
	dbal.On("GetExistingComputeTaskKeys", []string{}).Once().Return([]string{}, nil)

	dataManagerKey := newTrainTask.Data.(*asset.NewComputeTask_Train).Train.DataManagerKey
	dataSampleKeys := newTrainTask.Data.(*asset.NewComputeTask_Train).Train.DataSampleKeys

	dataManager := &asset.DataManager{
		Key:   dataManagerKey,
		Owner: "dm-owner",
		Permissions: &asset.Permissions{
			Process:  &asset.Permission{Public: true},
			Download: &asset.Permission{Public: true},
		},
		LogsPermission: &asset.Permission{Public: true},
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
	ps.On("IntersectPermissions", algo.Permissions, dataManager.Permissions).Return(modelPerms)

	storedTask := &asset.ComputeTask{
		Key:            newTrainTask.Key,
		Category:       newTrainTask.Category,
		Algo:           algo,
		Owner:          "testOwner",
		ComputePlanKey: newTrainTask.ComputePlanKey,
		Metadata:       newTrainTask.Metadata,
		Status:         asset.ComputeTaskStatus_STATUS_TODO,
		ParentTaskKeys: newTrainTask.ParentTaskKeys,
		Worker:         dataManager.Owner,
		Data: &asset.ComputeTask_Train{
			Train: &asset.TrainTaskData{
				DataManagerKey:   dataManagerKey,
				DataSampleKeys:   dataSampleKeys,
				ModelPermissions: modelPerms,
			},
		},
		CreationDate:   timestamppb.New(time.Unix(1337, 0)),
		LogsPermission: dataManager.LogsPermission,
	}

	// finally store the created task
	dbal.On("AddComputeTasks", storedTask).Once().Return(nil)

	expectedEvent := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
		AssetKey:  newTrainTask.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Metadata: map[string]string{
			"status": storedTask.Status.String(),
			"worker": dataManager.Owner,
		},
	}
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)

	err := service.RegisterTasks([]*asset.NewComputeTask{newTrainTask}, "testOwner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterFailedTask(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

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

	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	dbal.On("GetExistingComputeTaskKeys", newTask.ParentTaskKeys).Once().Return([]string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"}, nil)
	// Checking existing task
	dbal.On("ComputeTaskExists", newTask.Key).Once().Return(false, nil)

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

	err := service.RegisterTasks([]*asset.NewComputeTask{newTask}, "testOwner")
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrIncompatibleTaskStatus, orcError.Kind)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterDeletedModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	ms := new(MockModelAPI)
	provider := newMockedProvider()

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

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetModelService").Return(ms)

	service := NewComputeTaskService(provider)

	dbal.On("GetExistingComputeTaskKeys", newTask.ParentTaskKeys).Once().Return([]string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"}, nil)
	// Checking existing task
	dbal.On("ComputeTaskExists", newTask.Key).Once().Return(false, nil)

	parentPerms := &asset.Permissions{Process: &asset.Permission{Public: true}}
	parentTask := &asset.ComputeTask{
		Status:         asset.ComputeTaskStatus_STATUS_DONE,
		Key:            "6c3878a8-8ca6-437e-83be-3a85b24b70d1",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db82309fff",
		Data: &asset.ComputeTask_Train{
			Train: &asset.TrainTaskData{
				ModelPermissions: parentPerms,
			},
		},
	}

	dbal.On("GetComputeTasks", []string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"}).Once().
		Return([]*asset.ComputeTask{parentTask}, nil)

	ms.On("GetComputeTaskOutputModels", parentTask.Key).Once().Return([]*asset.Model{
		{Key: "uuid1", Address: &asset.Addressable{}},
		{Key: "disabled"},
	}, nil)

	err := service.RegisterTasks([]*asset.NewComputeTask{newTask}, "testOwner")
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrInvalidAsset, orcError.Kind)

	dbal.AssertExpectations(t)
	ms.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestSetCompositeData(t *testing.T) {
	taskInput := &asset.NewComputeTask{
		AlgoKey: "algoUuid",
	}
	specificInput := &asset.NewCompositeTrainTaskData{
		DataManagerKey: "dmUuid",
		DataSampleKeys: []string{"ds1", "ds2", "ds3"},
		TrunkPermissions: &asset.NewPermissions{
			Public:        false,
			AuthorizedIds: []string{"org1", "org2", "org3"},
		},
	}
	task := &asset.ComputeTask{
		Owner:    "org1",
		Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
	}

	dms := new(MockDataManagerAPI)
	dss := new(MockDataSampleAPI)
	ps := new(MockPermissionAPI)
	as := new(MockAlgoAPI)
	provider := newMockedProvider()
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetDataSampleService").Return(dss)
	provider.On("GetPermissionService").Return(ps)
	provider.On("GetAlgoService").Return(as)

	// getCheckedDataManager
	dms.On("GetDataManager", "dmUuid").Once().Return(&asset.DataManager{Key: "dmUuid", Owner: "dmOwner"}, nil)
	ps.On("CanProcess", mock.Anything, "org1").Return(true)
	dss.On("CheckSameManager", specificInput.DataManagerKey, specificInput.DataSampleKeys).Once().Return(nil)

	dss.On("ContainsTestSample", specificInput.DataSampleKeys).Once().Return(false, nil)

	// getCheckedAlgo
	algo := &asset.Algo{Category: asset.AlgoCategory_ALGO_COMPOSITE}
	as.On("GetAlgo", taskInput.AlgoKey).Once().Return(algo, nil)

	// create perms for head
	headPerms := &asset.Permissions{Process: &asset.Permission{Public: false, AuthorizedIds: []string{"dmOwner"}}}
	ps.On("CreatePermissions", "dmOwner", (*asset.NewPermissions)(nil)).Once().Return(headPerms, nil)
	// and trunk
	trunkPerms := &asset.Permissions{
		Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"dmOwner"}},
		Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "dmOwner"}},
	}
	ps.On("CreatePermissions", "dmOwner", specificInput.TrunkPermissions).Once().Return(trunkPerms, nil)

	service := NewComputeTaskService(provider)

	err := service.setCompositeData(taskInput, specificInput, task)
	assert.NoError(t, err)

	assert.Equal(t, algo, task.Algo)
	assert.Equal(t, "dmOwner", task.Worker)
	assert.Equal(t, "dmUuid", task.Data.(*asset.ComputeTask_Composite).Composite.DataManagerKey)
	assert.Equal(t, trunkPerms, task.Data.(*asset.ComputeTask_Composite).Composite.TrunkPermissions)
	assert.Equal(t, headPerms, task.Data.(*asset.ComputeTask_Composite).Composite.HeadPermissions)

	dms.AssertExpectations(t)
	dss.AssertExpectations(t)
	ps.AssertExpectations(t)
	as.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestSetAggregateData(t *testing.T) {
	ns := new(MockNodeAPI)
	as := new(MockAlgoAPI)
	provider := newMockedProvider()
	provider.On("GetNodeService").Return(ns)
	provider.On("GetAlgoService").Return(as)
	// Use the real permission service
	provider.On("GetPermissionService").Return(NewPermissionService(provider))

	taskInput := &asset.NewComputeTask{
		AlgoKey: "algoUuid",
	}
	specificInput := &asset.NewAggregateTrainTaskData{
		Worker: "org3",
	}
	task := &asset.ComputeTask{
		Owner:    "org1",
		Category: asset.ComputeTaskCategory_TASK_AGGREGATE,
	}

	parents := []*asset.ComputeTask{
		{
			Data: &asset.ComputeTask_Composite{
				Composite: &asset.CompositeTrainTaskData{
					TrunkPermissions: &asset.Permissions{
						Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org3"}},
						Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
					},
					HeadPermissions: &asset.Permissions{
						Process: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
					},
				},
			},
			LogsPermission: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
		},
		{
			Data: &asset.ComputeTask_Train{
				Train: &asset.TrainTaskData{
					ModelPermissions: &asset.Permissions{
						Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org4"}},
						Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org4"}},
					},
				},
			},
			LogsPermission: &asset.Permission{Public: false, AuthorizedIds: []string{"org4"}},
		},
	}

	// check node existence
	ns.On("GetNode", "org3").Once().Return(&asset.Node{Id: "org3"}, nil)
	// used by permissions service
	ns.On("GetAllNodes").Twice().Return([]*asset.Node{{Id: "org1"}, {Id: "org2"}, {Id: "org3"}}, nil)

	// getCheckedAlgo
	algo := &asset.Algo{Category: asset.AlgoCategory_ALGO_AGGREGATE, Permissions: &asset.Permissions{
		Process: &asset.Permission{Public: true},
	}}
	as.On("GetAlgo", taskInput.AlgoKey).Once().Return(algo, nil)

	service := NewComputeTaskService(provider)
	err := service.setAggregateData(taskInput, specificInput, task, parents)

	assert.NoError(t, err)

	assert.Equal(t, algo, task.Algo)
	assert.Equal(t, "org3", task.Worker)
	assert.False(t, task.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions.Process.Public)
	assert.False(t, task.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions.Download.Public)
	assert.ElementsMatch(t, task.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions.Process.AuthorizedIds, []string{"org1", "org3", "org4"})
	assert.ElementsMatch(t, task.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions.Download.AuthorizedIds, []string{"org1", "org4"})
	assert.ElementsMatch(t, task.LogsPermission.AuthorizedIds, []string{"org1", "org2", "org4"})

	ns.AssertExpectations(t)
	as.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestSetTestData(t *testing.T) {
	t.Run("single metric", func(t *testing.T) {
		specificInput := &asset.NewTestTaskData{
			MetricKeys:     []string{"metric"},
			DataManagerKey: "cdmKey",
			DataSampleKeys: []string{"sample1", "sample2"},
		}
		task := &asset.ComputeTask{
			Owner:    "org1",
			Category: asset.ComputeTaskCategory_TASK_TEST,
		}
		parents := []*asset.ComputeTask{
			{
				Algo:           &asset.Algo{Key: "algoKey"},
				ComputePlanKey: "cpKey",
				Rank:           2,
			},
		}

		dbal := new(persistence.MockDBAL)
		ms := new(MockMetricAPI)
		dms := new(MockDataManagerAPI)
		dss := new(MockDataSampleAPI)
		provider := newMockedProvider()
		provider.On("GetMetricDBAL").Return(dbal)
		provider.On("GetMetricService").Return(ms)
		provider.On("GetDataManagerService").Return(dms)
		provider.On("GetDataSampleService").Return(dss)
		provider.On("GetPermissionService").Return(NewPermissionService(provider))
		service := NewComputeTaskService(provider)

		// single metric
		dbal.On("MetricExists", "metric").Return(true, nil)
		ms.On("CanDownload", "metric", "dmowner").Return(true, nil)
		dms.On("GetDataManager", "cdmKey").Once().Return(&asset.DataManager{Key: "cdmKey", Permissions: &asset.Permissions{Process: &asset.Permission{Public: true}}, Owner: "dmowner"}, nil)
		dss.On("CheckSameManager", specificInput.DataManagerKey, specificInput.DataSampleKeys).Once().Return(nil)

		err := service.setTestData(specificInput, task, parents)
		assert.NoError(t, err)
		assert.Equal(t, parents[0].Algo, task.Algo)
		assert.Equal(t, parents[0].ComputePlanKey, task.ComputePlanKey)
		assert.Equal(t, int32(2), task.Rank, "test task should have the same rank than its parent")
		assert.Equal(t, task.Data.(*asset.ComputeTask_Test).Test.DataManagerKey, specificInput.DataManagerKey)
		assert.Equal(t, task.Data.(*asset.ComputeTask_Test).Test.DataSampleKeys, specificInput.DataSampleKeys)
		assert.Equal(t, task.Data.(*asset.ComputeTask_Test).Test.MetricKeys, specificInput.MetricKeys)

		dbal.AssertExpectations(t)
		ms.AssertExpectations(t)
		provider.AssertExpectations(t)
	})
	t.Run("multiple metrics", func(t *testing.T) {
		specificInput := &asset.NewTestTaskData{
			MetricKeys:     []string{"metric1", "metric2"},
			DataManagerKey: "cdmKey",
			DataSampleKeys: []string{"sample1", "sample2"},
		}
		task := &asset.ComputeTask{
			Owner:    "org1",
			Category: asset.ComputeTaskCategory_TASK_TEST,
		}
		parents := []*asset.ComputeTask{
			{
				Algo:           &asset.Algo{Key: "algoKey"},
				ComputePlanKey: "cpKey",
				Rank:           2,
			},
		}

		dbal := new(persistence.MockDBAL)
		ms := new(MockMetricAPI)
		dms := new(MockDataManagerAPI)
		dss := new(MockDataSampleAPI)
		provider := newMockedProvider()
		provider.On("GetMetricDBAL").Return(dbal)
		provider.On("GetMetricService").Return(ms)
		provider.On("GetDataManagerService").Return(dms)
		provider.On("GetDataSampleService").Return(dss)
		provider.On("GetPermissionService").Return(NewPermissionService(provider))
		service := NewComputeTaskService(provider)

		// multiple metrics
		dbal.On("MetricExists", "metric1").Return(true, nil)
		dbal.On("MetricExists", "metric2").Return(true, nil)
		ms.On("CanDownload", "metric1", "dmowner").Return(true, nil)
		ms.On("CanDownload", "metric2", "dmowner").Return(true, nil)
		dms.On("GetDataManager", "cdmKey").Once().Return(&asset.DataManager{Key: "cdmKey", Permissions: &asset.Permissions{Process: &asset.Permission{Public: true}}, Owner: "dmowner"}, nil)
		dss.On("CheckSameManager", specificInput.DataManagerKey, specificInput.DataSampleKeys).Once().Return(nil)

		err := service.setTestData(specificInput, task, parents)
		assert.NoError(t, err)
		assert.Equal(t, parents[0].Algo, task.Algo)
		assert.Equal(t, parents[0].ComputePlanKey, task.ComputePlanKey)
		assert.Equal(t, int32(2), task.Rank, "test task should have the same rank than its parent")
		assert.Equal(t, task.Data.(*asset.ComputeTask_Test).Test.DataManagerKey, specificInput.DataManagerKey)
		assert.Equal(t, task.Data.(*asset.ComputeTask_Test).Test.DataSampleKeys, specificInput.DataSampleKeys)
		assert.Equal(t, task.Data.(*asset.ComputeTask_Test).Test.MetricKeys, specificInput.MetricKeys)

		dbal.AssertExpectations(t)
		ms.AssertExpectations(t)
		provider.AssertExpectations(t)
	})
	t.Run("invalid metric", func(t *testing.T) {
		specificInput := &asset.NewTestTaskData{
			MetricKeys:     []string{"metric"},
			DataManagerKey: "cdmKey",
			DataSampleKeys: []string{"sample1", "sample2"},
		}
		task := &asset.ComputeTask{
			Owner:    "org1",
			Category: asset.ComputeTaskCategory_TASK_TEST,
		}
		parents := []*asset.ComputeTask{
			{
				Algo:           &asset.Algo{Key: "algoKey"},
				ComputePlanKey: "cpKey",
				Rank:           2,
			},
		}

		dbal := new(persistence.MockDBAL)
		ms := new(MockMetricAPI)
		dms := new(MockDataManagerAPI)
		dss := new(MockDataSampleAPI)
		provider := newMockedProvider()
		provider.On("GetMetricDBAL").Return(dbal)
		provider.On("GetMetricService").Return(ms)
		provider.On("GetDataManagerService").Return(dms)
		provider.On("GetDataSampleService").Return(dss)
		// Use the real permission service
		provider.On("GetPermissionService").Return(NewPermissionService(provider))
		service := NewComputeTaskService(provider)

		// can not download metric
		dbal.On("MetricExists", "metric").Return(true, nil)
		ms.On("CanDownload", "metric", "dmowner").Return(false, nil)
		dms.On("GetDataManager", "cdmKey").Once().Return(&asset.DataManager{Key: "cdmKey", Permissions: &asset.Permissions{Process: &asset.Permission{Public: true}}, Owner: "dmowner"}, nil)
		dss.On("CheckSameManager", specificInput.DataManagerKey, specificInput.DataSampleKeys).Once().Return(nil)

		err := service.setTestData(specificInput, task, parents)
		assert.Error(t, err)

		dbal.AssertExpectations(t)
		ms.AssertExpectations(t)
		provider.AssertExpectations(t)
	})
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

	provider := newMockedProvider()
	service := NewComputeTaskService(provider)

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("%s: %t", c.n, c.o),
			func(t *testing.T) {
				assert.Equal(t, c.o, service.isCompatibleWithParents(c.t, c.p))
			},
		)
	}
}

func createNode(parent string, key string) *asset.NewComputeTask {
	if parent != "" {
		return &asset.NewComputeTask{
			Key:            key,
			ParentTaskKeys: []string{parent},
		}
	}

	return &asset.NewComputeTask{
		Key:            key,
		ParentTaskKeys: []string{},
	}
}

func TestSortTasks(t *testing.T) {
	//     +-->Leaf1
	//     |
	//     |
	// Root|                  +->Leaf5
	//     |        +-->Node3-+
	//     |        |
	//     +-->Node2|
	//              |
	//              +-->Leaf4

	root := &asset.NewComputeTask{
		Key:            "root",
		ParentTaskKeys: []string{},
	}

	leaf1 := createNode(root.Key, "leaf1")
	node2 := createNode(root.Key, "node2")
	node3 := createNode(node2.Key, "node3")
	leaf4 := createNode(node2.Key, "leaf4")
	leaf5 := createNode(node3.Key, "leaf5")

	nodes := []*asset.NewComputeTask{root, leaf5, leaf4, node2, node3, leaf1}
	existingKeys := []string{}

	provider := newMockedProvider()
	service := NewComputeTaskService(provider)
	result, err := service.SortTasks(nodes, existingKeys)

	assert.NoError(t, err)
	assert.Equal(t, len(nodes), len(result))
	assert.ElementsMatch(t, nodes, result)
	assert.Equal(t, root, result[0])
}

func TestSortTasksWithCircularDependency(t *testing.T) {
	//     +-->Leaf1
	//     |
	//     |
	// Root|                  +->Leaf5
	//     |        +-->Node3-+
	//     |        |
	//     +-->Node2|
	//           ^  |
	//           |  +-->Leaf4-+
	//           |            |
	//           +------------+

	root := &asset.NewComputeTask{
		Key:            "root",
		ParentTaskKeys: []string{},
	}

	leaf1 := createNode(root.Key, "leaf1")
	node2 := createNode(root.Key, "node2")
	node3 := createNode(node2.Key, "node3")
	leaf4 := createNode(node2.Key, "leaf4")
	leaf5 := createNode(node3.Key, "leaf5")

	node2.ParentTaskKeys = []string{root.Key, leaf4.Key}

	nodes := []*asset.NewComputeTask{root, leaf5, leaf4, node2, node3, leaf1}
	existingKeys := []string{}

	provider := newMockedProvider()
	service := NewComputeTaskService(provider)
	_, err := service.SortTasks(nodes, existingKeys)

	assert.Error(t, err)
}

func TestSortDependencyWithExistingTasks(t *testing.T) {
	//        Existing2-----+-->Leaf1
	//                      |
	//                      |
	// Existing1------->Root|                  +->Leaf5
	//                      |        +-->Node3-+
	//                      |        |
	//                      +-->Node2|
	//                               |
	//                               +-->Leaf4

	existing1 := "exist1"
	existing2 := "exist2"

	root := &asset.NewComputeTask{
		Key:            "root",
		ParentTaskKeys: []string{existing1},
	}

	leaf1 := createNode(root.Key, "leaf1")
	node2 := createNode(root.Key, "node2")
	node3 := createNode(node2.Key, "node3")
	leaf4 := createNode(node2.Key, "leaf4")
	leaf5 := createNode(node3.Key, "leaf5")

	leaf1.ParentTaskKeys = []string{existing2, root.Key}

	nodes := []*asset.NewComputeTask{root, leaf5, leaf4, node2, node3, leaf1}
	existingKeys := []string{existing1, existing2}

	provider := newMockedProvider()
	service := NewComputeTaskService(provider)
	result, err := service.SortTasks(nodes, existingKeys)

	assert.NoError(t, err)
	assert.Equal(t, len(nodes), len(result))
	assert.ElementsMatch(t, nodes, result)
	assert.Equal(t, root, result[0])
}

func TestSortTasksUnknownRef(t *testing.T) {
	//         +-->Leaf1
	//         |
	//         |
	// ?-->Root|                  +->Leaf5
	//         |        +-->Node3-+
	//         |        |
	//         +-->Node2|
	//                  |
	//                  +-->Leaf4

	root := &asset.NewComputeTask{
		Key:            "root",
		ParentTaskKeys: []string{"unknown"},
	}

	leaf1 := createNode(root.Key, "leaf1")
	node2 := createNode(root.Key, "node2")
	node3 := createNode(node2.Key, "node3")
	leaf4 := createNode(node2.Key, "leaf4")
	leaf5 := createNode(node3.Key, "leaf5")

	nodes := []*asset.NewComputeTask{root, leaf5, leaf4, node2, node3, leaf1}
	existingKeys := []string{}

	provider := newMockedProvider()
	service := NewComputeTaskService(provider)
	_, err := service.SortTasks(nodes, existingKeys)

	assert.Error(t, err)
}

func TestCheckCanProcessParent(t *testing.T) {
	parents := []*asset.ComputeTask{
		{
			Data: &asset.ComputeTask_Train{
				Train: &asset.TrainTaskData{
					ModelPermissions: &asset.Permissions{
						Process: &asset.Permission{Public: true},
					},
				},
			},
		},
		{
			Data: &asset.ComputeTask_Composite{
				Composite: &asset.CompositeTrainTaskData{
					TrunkPermissions: &asset.Permissions{
						Process: &asset.Permission{Public: false, AuthorizedIds: []string{"orgTest", "org2"}},
					},
					HeadPermissions: &asset.Permissions{
						Process: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
					},
				},
			},
		},
		{
			Data: &asset.ComputeTask_Aggregate{
				Aggregate: &asset.AggregateTrainTaskData{
					ModelPermissions: &asset.Permissions{
						Process: &asset.Permission{Public: false, AuthorizedIds: []string{"orgTest", "org2"}},
					},
				},
			},
		},
	}

	cases := map[string]struct {
		requester    string
		taskCategory asset.ComputeTaskCategory
		canProcess   bool
	}{
		"train task": {
			"orgTest",
			asset.ComputeTaskCategory_TASK_TRAIN,
			true,
		},
		"test task": {
			"orgTest",
			asset.ComputeTaskCategory_TASK_TEST,
			false, // cannot test head from parent composite
		},
		"aggregate task": {
			"org2",
			asset.ComputeTaskCategory_TASK_AGGREGATE,
			true,
		},
	}
	provider := newMockedProvider()
	// Use the real permission service
	provider.On("GetPermissionService").Return(NewPermissionService(provider))
	service := NewComputeTaskService(provider)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := service.checkCanProcessParents(tc.requester, parents, tc.taskCategory)

			if tc.canProcess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				orcError := new(orcerrors.OrcError)
				assert.True(t, errors.As(err, &orcError))
				assert.Equal(t, orcerrors.ErrPermissionDenied, orcError.Kind)
			}
		})
	}
}

func TestGetRank(t *testing.T) {
	parents := []*asset.ComputeTask{
		{Rank: 0},
		{Rank: 1},
		{Rank: 3},
	}
	assert.Equal(t, int32(4), getRank(parents))

	noParents := []*asset.ComputeTask{}
	assert.Equal(t, int32(0), getRank(noParents))
}

func TestCanDisableModels(t *testing.T) {
	t.Run("not worker", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DOING,
			Worker: "woker",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)

		service := NewComputeTaskService(provider)
		_, err := service.canDisableModels("uuid", "notworker")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrPermissionDenied, orcError.Kind)

		dbal.AssertExpectations(t)
	})
	t.Run("non terminal task", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DOING,
			Worker: "worker",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)

		service := NewComputeTaskService(provider)
		can, err := service.canDisableModels("uuid", "worker")
		assert.NoError(t, err)
		assert.False(t, can)

		dbal.AssertExpectations(t)
	})
	t.Run("compute plan cannot be disabled", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status:         asset.ComputeTaskStatus_STATUS_DONE,
			ComputePlanKey: "cpKey",
			Worker:         "worker",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		cps := new(MockComputePlanAPI)
		provider.On("GetComputePlanService").Return(cps)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)

		cps.On("canDeleteModels", "cpKey").Return(false, nil)

		service := NewComputeTaskService(provider)
		can, err := service.canDisableModels("uuid", "worker")
		assert.NoError(t, err)
		assert.False(t, can)

		dbal.AssertExpectations(t)
		cps.AssertExpectations(t)
	})
	t.Run("task without children", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status:         asset.ComputeTaskStatus_STATUS_DONE,
			ComputePlanKey: "cpKey",
			Worker:         "worker",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		cps := new(MockComputePlanAPI)
		provider.On("GetComputePlanService").Return(cps)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		cps.On("canDeleteModels", "cpKey").Return(true, nil)
		dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{}, nil)

		service := NewComputeTaskService(provider)
		can, err := service.canDisableModels("uuid", "worker")
		assert.NoError(t, err)
		assert.False(t, can)

		dbal.AssertExpectations(t)
		cps.AssertExpectations(t)
	})
	t.Run("task with active children", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status:         asset.ComputeTaskStatus_STATUS_DONE,
			ComputePlanKey: "cpKey",
			Worker:         "worker",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		cps := new(MockComputePlanAPI)
		provider.On("GetComputePlanService").Return(cps)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		cps.On("canDeleteModels", "cpKey").Return(true, nil)
		dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{
			{Status: asset.ComputeTaskStatus_STATUS_DOING},
		}, nil)

		service := NewComputeTaskService(provider)
		can, err := service.canDisableModels("uuid", "worker")
		assert.NoError(t, err)
		assert.False(t, can)

		dbal.AssertExpectations(t)
		cps.AssertExpectations(t)
	})
	t.Run("task with only test children", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status:         asset.ComputeTaskStatus_STATUS_DONE,
			ComputePlanKey: "cpKey",
			Worker:         "worker",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		cps := new(MockComputePlanAPI)
		provider.On("GetComputePlanService").Return(cps)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		cps.On("canDeleteModels", "cpKey").Return(true, nil)
		dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{
			{Status: asset.ComputeTaskStatus_STATUS_DONE, Category: asset.ComputeTaskCategory_TASK_TEST},
		}, nil)

		service := NewComputeTaskService(provider)
		can, err := service.canDisableModels("uuid", "worker")
		assert.NoError(t, err)
		assert.False(t, can)

		dbal.AssertExpectations(t)
		cps.AssertExpectations(t)
	})
	t.Run("task can be disabled", func(t *testing.T) {
		task := &asset.ComputeTask{
			Key:            "uuid",
			Status:         asset.ComputeTaskStatus_STATUS_DONE,
			ComputePlanKey: "cpKey",
			Worker:         "worker",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		cps := new(MockComputePlanAPI)
		provider.On("GetComputePlanService").Return(cps)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		cps.On("canDeleteModels", "cpKey").Return(true, nil)
		dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{
			{Status: asset.ComputeTaskStatus_STATUS_DONE},
		}, nil)

		service := NewComputeTaskService(provider)
		can, err := service.canDisableModels("uuid", "worker")
		assert.NoError(t, err)
		assert.True(t, can)

		dbal.AssertExpectations(t)
		cps.AssertExpectations(t)
	})
}

func TestRegisterTasksEmptyList(t *testing.T) {
	provider := newMockedProvider()

	service := NewComputeTaskService(provider)

	err := service.RegisterTasks([]*asset.NewComputeTask{}, "test")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrBadRequest, orcError.Kind)
}

func TestGetRegisteredTask(t *testing.T) {
	provider := newMockedProvider()
	dbal := new(persistence.MockDBAL)
	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	// simulate a task in store
	service.taskStore["uuid1"] = &asset.ComputeTask{Key: "uuid1"}
	// Then, there should be only one task to retrieve from DB
	dbal.On("GetComputeTasks", []string{"uuid2"}).Once().Return(
		[]*asset.ComputeTask{{Key: "uuid2"}},
		nil,
	)

	tasks, err := service.getRegisteredTasks("uuid1", "uuid2")
	assert.NoError(t, err)

	assert.Len(t, tasks, 2)

	dbal.AssertExpectations(t)
}
