package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var newPerms = &asset.NewPermissions{
	AuthorizedIds: []string{"testOwner"},
}

var (
	dataManagerKey = "2837f0b7-cb0e-4a98-9df2-68c116f65ad6"
	dataSampleKeys = []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"}
)

var newTrainTask = &asset.NewComputeTask{
	Key:            "867852b4-8419-4d52-8862-d5db823095be",
	FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
	ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
	Inputs: []*asset.ComputeTaskInput{
		{Identifier: "data", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: dataSampleKeys[0]}},
		{Identifier: "data", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: dataSampleKeys[1]}},
		{Identifier: "opener", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: dataManagerKey}},
	},
	Outputs: map[string]*asset.NewComputeTaskOutput{
		"model": {
			Permissions: newPerms,
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
		Status: asset.ComputeTaskStatus_STATUS_EXECUTING,
	}

	returnedTasks := []*asset.ComputeTask{{}, {}}

	dbal.On("QueryComputeTasks", pagination, filter).Once().Return(returnedTasks, "", nil)

	tasks, _, err := service.QueryTasks(pagination, filter)
	assert.NoError(t, err)

	assert.Len(t, tasks, 2)
}

func TestRegisterMissingComputePlan(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cps := new(MockComputePlanAPI)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetComputePlanService").Return(cps)

	service := NewComputeTaskService(provider)

	dbal.On("GetExistingComputeTaskKeys", []string{newTrainTask.Key}).Once().Return([]string{}, nil)
	dbal.On("GetExistingComputeTaskKeys", []string{}).Once().Return([]string{}, nil)
	cps.On("GetPlan", newTrainTask.ComputePlanKey).Once().Return(nil, orcerrors.NewNotFound("compute plan", newTrainTask.ComputePlanKey))

	_, err := service.RegisterTasks([]*asset.NewComputeTask{newTrainTask}, "test")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrNotFound, orcError.Kind)

	dbal.AssertExpectations(t)
	cps.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterTasksComputePlanOwnedBySomeoneElse(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cps := new(MockComputePlanAPI)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetComputePlanService").Return(cps)

	service := NewComputeTaskService(provider)

	dbal.On("GetExistingComputeTaskKeys", []string{newTrainTask.Key}).Once().Return([]string{}, nil)
	dbal.On("GetExistingComputeTaskKeys", []string{}).Once().Return([]string{}, nil)
	cps.On("GetPlan", newTrainTask.ComputePlanKey).Once().Return(&asset.ComputePlan{Key: newTrainTask.ComputePlanKey, Owner: "not test"}, nil)

	_, err := service.RegisterTasks([]*asset.NewComputeTask{newTrainTask}, "test")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrPermissionDenied, orcError.Kind)

	dbal.AssertExpectations(t)
	cps.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterTaskConflict(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputeTaskDBAL").Return(dbal)
	dbal.On("GetExistingComputeTaskKeys", []string{newTrainTask.Key}).Once().Return([]string{newTrainTask.Key}, nil)

	service := NewComputeTaskService(provider)
	_, err := service.RegisterTasks([]*asset.NewComputeTask{newTrainTask}, "test")
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

	cps := new(MockComputePlanAPI)
	dms := new(MockDataManagerAPI)
	ps := new(MockPermissionAPI)
	as := new(MockFunctionAPI)
	ts := new(MockTimeAPI)
	os := new(MockOrganizationAPI)

	provider.On("GetComputePlanService").Return(cps)
	provider.On("GetEventService").Return(es)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetPermissionService").Return(ps)
	provider.On("GetFunctionService").Return(as)
	provider.On("GetTimeService").Return(ts)
	provider.On("GetOrganizationService").Return(os)

	cps.On("GetPlan", newTrainTask.ComputePlanKey).Once().Return(&asset.ComputePlan{Key: newTrainTask.Key, Owner: "testOwner"}, nil)
	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	service := NewComputeTaskService(provider)

	// Checking existing task
	dbal.On("GetExistingComputeTaskKeys", []string{newTrainTask.Key}).Once().Return([]string{}, nil)
	dbal.On("GetExistingComputeTaskKeys", []string{}).Once().Return([]string{}, nil)

	dataManager := &asset.DataManager{
		Key:   dataManagerKey,
		Owner: "dm-owner",
		Permissions: &asset.Permissions{
			Process:  &asset.Permission{Public: true},
			Download: &asset.Permission{Public: true},
		},
		LogsPermission: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
	}
	os.On("GetOrganization", dataManager.Owner).Once().Return(&asset.Organization{Id: dataManager.Owner}, nil)

	// Checking datamanager permissions
	dms.On("GetDataManager", dataManagerKey).Once().Return(dataManager, nil)
	dms.On("CheckDataManager", dataManager, dataSampleKeys, "testOwner").Once().Return(nil)

	function := &asset.Function{
		Key:    "b09cc8eb-cb76-49ce-8f93-2f8b3185e7b7",
		Status: asset.FunctionStatus_FUNCTION_STATUS_READY,
		Permissions: &asset.Permissions{
			Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"testOwner"}},
			Download: &asset.Permission{Public: false, AuthorizedIds: []string{"testOwner"}},
		},
		Inputs: map[string]*asset.FunctionInput{
			"data":   {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"opener": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
		},
		Outputs: map[string]*asset.FunctionOutput{
			"model": {
				Kind: asset.AssetKind_ASSET_MODEL,
			},
		},
	}
	// check function matching
	as.On("GetFunction", newTrainTask.FunctionKey).Once().Return(function, nil)
	ps.On("CanProcess", function.Permissions, "testOwner").Once().Return(true)

	// create new permissions
	modelPerms := &asset.Permissions{
		Process:  &asset.Permission{AuthorizedIds: []string{"testOwner"}},
		Download: &asset.Permission{AuthorizedIds: []string{"testOwner"}},
	}
	ps.On("CreatePermissions", "testOwner", newPerms).Return(modelPerms, nil)

	storedTask := &asset.ComputeTask{
		Key:            newTrainTask.Key,
		FunctionKey:    function.Key,
		Owner:          "testOwner",
		ComputePlanKey: newTrainTask.ComputePlanKey,
		Metadata:       newTrainTask.Metadata,
		Status:         asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT,
		Worker:         dataManager.Owner,
		Inputs:         newTrainTask.Inputs,
		CreationDate:   timestamppb.New(time.Unix(1337, 0)),
		LogsPermission: dataManager.LogsPermission,
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {Permissions: modelPerms},
		},
	}

	// finally store the created task
	dbal.On("AddComputeTasks", storedTask).Once().Return(nil)

	expectedEvent := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
		AssetKey:  newTrainTask.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_ComputeTask{ComputeTask: storedTask},
	}
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)

	_, err := service.RegisterTasks([]*asset.NewComputeTask{newTrainTask}, "testOwner")
	assert.NoError(t, err)

	cps.AssertExpectations(t)
	dms.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
	as.AssertExpectations(t)
	os.AssertExpectations(t)
	ps.AssertExpectations(t)
}

func TestRegisterCompositeTaskWithCompositeParents(t *testing.T) {
	sharedPermsNew := &asset.NewPermissions{
		AuthorizedIds: []string{"testOwner", "otherOrg"},
	}
	sharedPerms := &asset.Permissions{
		Process:  &asset.Permission{AuthorizedIds: []string{"testOwner"}},
		Download: &asset.Permission{AuthorizedIds: []string{"testOwner"}},
	}

	localPermsNew := &asset.NewPermissions{
		AuthorizedIds: []string{"testOwner"},
	}
	localPerms := &asset.Permissions{
		Process:  &asset.Permission{AuthorizedIds: []string{"testOwner"}},
		Download: &asset.Permission{AuthorizedIds: []string{"testOwner"}},
	}

	newTask := &asset.NewComputeTask{
		Key:            "aaaaaaaa-cccc-bbbb-eeee-ffffffffffff",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*asset.ComputeTaskInput{
			{Identifier: "local", Ref: &asset.ComputeTaskInput_ParentTaskOutput{
				ParentTaskOutput: &asset.ParentTaskOutputRef{
					OutputIdentifier: "local",
					ParentTaskKey:    "aaaaaaaa-cccc-bbbb-eeee-111111111111",
				},
			}},
			{Identifier: "shared", Ref: &asset.ComputeTaskInput_ParentTaskOutput{
				ParentTaskOutput: &asset.ParentTaskOutputRef{
					OutputIdentifier: "shared",
					ParentTaskKey:    "aaaaaaaa-cccc-bbbb-eeee-222222222222",
				},
			}},
			{Identifier: "opener", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: dataManagerKey}},
			{Identifier: "data", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: dataSampleKeys[0]}},
			{Identifier: "data", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: dataSampleKeys[1]}},
		},
		Outputs: map[string]*asset.NewComputeTaskOutput{
			"shared": {Permissions: sharedPermsNew},
			"local":  {Permissions: localPermsNew},
		},
	}

	permissions := &asset.Permissions{
		Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"testOwner"}},
		Download: &asset.Permission{Public: false, AuthorizedIds: []string{"testOwner"}},
	}

	functionParent1 := &asset.Function{
		Key:    "d29118a9-f989-41af-ae02-90f0c4aaffe3",
		Status: asset.FunctionStatus_FUNCTION_STATUS_READY,
		Outputs: map[string]*asset.FunctionOutput{
			"local":  {Kind: asset.AssetKind_ASSET_MODEL},
			"shared": {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
	functionParent2 := &asset.Function{
		Key:    "cc765417-1e14-41c8-9f7b-653ed335d30d",
		Status: asset.FunctionStatus_FUNCTION_STATUS_READY,
		Outputs: map[string]*asset.FunctionOutput{
			"local":  {Kind: asset.AssetKind_ASSET_MODEL},
			"shared": {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}

	parent1 := &asset.ComputeTask{
		Key:            "aaaaaaaa-cccc-bbbb-eeee-111111111111",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Status:         asset.ComputeTaskStatus_STATUS_EXECUTING,
		FunctionKey:    functionParent1.Key,
		Outputs: map[string]*asset.ComputeTaskOutput{
			"shared": {Permissions: sharedPerms},
			"local":  {Permissions: localPerms},
		},
	}
	parent2 := &asset.ComputeTask{
		Key:            "aaaaaaaa-cccc-bbbb-eeee-222222222222",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Status:         asset.ComputeTaskStatus_STATUS_EXECUTING,
		FunctionKey:    functionParent2.Key,
		Outputs: map[string]*asset.ComputeTaskOutput{
			"shared": {Permissions: sharedPerms},
			"local":  {Permissions: localPerms},
		},
	}

	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	cps := new(MockComputePlanAPI)
	dms := new(MockDataManagerAPI)
	ps := new(MockPermissionAPI)
	as := new(MockFunctionAPI)
	ts := new(MockTimeAPI)
	os := new(MockOrganizationAPI)

	provider.On("GetComputePlanService").Return(cps)
	provider.On("GetEventService").Return(es)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetPermissionService").Return(ps)
	provider.On("GetFunctionService").Return(as)
	provider.On("GetTimeService").Return(ts)
	provider.On("GetOrganizationService").Return(os)

	cps.On("GetPlan", newTask.ComputePlanKey).Once().Return(&asset.ComputePlan{Key: newTask.ComputePlanKey, Owner: "testOwner"}, nil)
	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	service := NewComputeTaskService(provider)

	// Checking existing task
	dbal.On("GetExistingComputeTaskKeys", []string{newTask.Key}).Once().Return([]string{}, nil)
	// All parents already exist
	dbal.On("GetExistingComputeTaskKeys", []string{parent1.Key, parent2.Key}).Once().Return([]string{parent1.Key, parent2.Key}, nil)

	// STATUS_WAITING_FOR_EXECUTOR_SLOT: we fetch the same data several times
	// Since this will change with task category removal, let's revisit later
	dbal.On("GetComputeTasks", []string{parent1.Key, parent2.Key}).Once().Return([]*asset.ComputeTask{parent1, parent2}, nil)
	dbal.On("GetComputeTask", parent1.Key).Once().Return(parent1, nil)
	dbal.On("GetComputeTask", parent2.Key).Once().Return(parent2, nil)

	dataManager := &asset.DataManager{
		Key:   dataManagerKey,
		Owner: "dm-owner",
		Permissions: &asset.Permissions{
			Process:  &asset.Permission{Public: true},
			Download: &asset.Permission{Public: true},
		},
	}
	os.On("GetOrganization", dataManager.Owner).Once().Return(&asset.Organization{Id: dataManager.Owner}, nil)

	// Checking datamanager permissions
	dms.On("GetDataManager", dataManagerKey).Once().Return(dataManager, nil)
	// Checked twice while we still deal with task specific data
	dms.On("CheckDataManager", dataManager, dataSampleKeys, "testOwner").Once().Return(nil)

	// create permissions
	ps.On("CreatePermissions", "testOwner", sharedPermsNew).Return(sharedPerms, nil)
	ps.On("CreatePermissions", "testOwner", localPermsNew).Return(localPerms, nil)

	function := &asset.Function{
		Permissions: permissions,
		Inputs: map[string]*asset.FunctionInput{
			"local":  {Kind: asset.AssetKind_ASSET_MODEL},
			"shared": {Kind: asset.AssetKind_ASSET_MODEL},
			"opener": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"data":   {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
		},
		Outputs: map[string]*asset.FunctionOutput{
			"shared": {Kind: asset.AssetKind_ASSET_MODEL},
			"local":  {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
	// check function matching
	as.On("GetFunction", newTask.FunctionKey).Once().Return(function, nil)
	ps.On("CanProcess", function.Permissions, "testOwner").Once().Return(true)

	// Parent processing check -> requester is the task worker, so the datamanager owner here
	as.On("GetFunction", parent1.FunctionKey).Once().Return(functionParent1, nil)
	ps.On("CanProcess", parent1.Outputs["local"].Permissions, dataManager.Owner).Once().Return(true)
	as.On("GetFunction", parent2.FunctionKey).Once().Return(functionParent2, nil)
	ps.On("CanProcess", parent2.Outputs["shared"].Permissions, dataManager.Owner).Once().Return(true)

	storedTask := &asset.ComputeTask{
		Key:            newTask.Key,
		FunctionKey:    function.Key,
		Owner:          "testOwner",
		ComputePlanKey: newTask.ComputePlanKey,
		Metadata:       newTask.Metadata,
		Worker:         dataManager.Owner,
		Rank:           1,
		CreationDate:   timestamppb.New(time.Unix(1337, 0)),
		Inputs:         newTask.Inputs,
		Outputs: map[string]*asset.ComputeTaskOutput{
			"shared": {Permissions: sharedPerms},
			"local":  {Permissions: localPerms},
		},
	}

	// finally store the created task
	dbal.On("AddComputeTasks", storedTask).Once().Return(nil)

	expectedEvent := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
		AssetKey:  newTask.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_ComputeTask{ComputeTask: storedTask},
	}
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)

	_, err := service.RegisterTasks([]*asset.NewComputeTask{newTask}, "testOwner")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	as.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
	ps.AssertExpectations(t)
	dms.AssertExpectations(t)
	cps.AssertExpectations(t)
}

func TestRegisterDeletedModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	ms := new(MockModelAPI)
	fs := new(MockFunctionAPI)
	ps := new(MockPermissionAPI)
	cps := new(MockComputePlanAPI)
	provider := newMockedProvider()

	newTask := &asset.NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*asset.ComputeTaskInput{
			{Identifier: "model", Ref: &asset.ComputeTaskInput_ParentTaskOutput{
				ParentTaskOutput: &asset.ParentTaskOutputRef{
					OutputIdentifier: "model",
					ParentTaskKey:    "6c3878a8-8ca6-437e-83be-3a85b24b70d1",
				},
			}},
		},
	}

	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetModelService").Return(ms)
	provider.On("GetComputePlanService").Return(cps)
	provider.On("GetFunctionService").Return(fs)
	provider.On("GetPermissionService").Return(ps)

	service := NewComputeTaskService(provider)

	cps.On("GetPlan", newTrainTask.ComputePlanKey).Once().Return(&asset.ComputePlan{Key: newTrainTask.ComputePlanKey, Owner: "testOwner"}, nil)

	function := &asset.Function{
		Status: asset.FunctionStatus_FUNCTION_STATUS_READY,
	}
	ps.On("CanProcess", function.Permissions, "testOwner").Return(true).Once()
	fs.On("GetFunction", newTask.Key).Return(function, nil).Once()

	// Checking existing task
	dbal.On("GetExistingComputeTaskKeys", []string{newTask.Key}).Once().Return([]string{}, nil)

	dbal.On("GetExistingComputeTaskKeys", []string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"}).Once().Return([]string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"}, nil)

	parentPerms := &asset.Permissions{Process: &asset.Permission{Public: true}}
	parentTask := &asset.ComputeTask{
		Status:         asset.ComputeTaskStatus_STATUS_DONE,
		Key:            "6c3878a8-8ca6-437e-83be-3a85b24b70d1",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db82309fff",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {Permissions: parentPerms},
		},
	}

	dbal.On("GetComputeTasks", []string{"6c3878a8-8ca6-437e-83be-3a85b24b70d1"}).Once().
		Return([]*asset.ComputeTask{parentTask}, nil)

	ms.On("GetComputeTaskOutputModels", parentTask.Key).Once().Return([]*asset.Model{
		{Key: "uuid1", Address: &asset.Addressable{}},
		{Key: "disabled"},
	}, nil)

	_, err := service.RegisterTasks([]*asset.NewComputeTask{newTask}, "testOwner")
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrInvalidAsset, orcError.Kind)

	dbal.AssertExpectations(t)
	cps.AssertExpectations(t)
	ms.AssertExpectations(t)
	fs.AssertExpectations(t)
	ps.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestValidateTaskInputs(t *testing.T) {

	type dependenciesErrors struct {
		getComputeTask   error
		checkDataManager error
		getCheckedModel  error
	}
	owner := "org1"
	defaultWorker := "org2"

	validRef := &asset.ComputeTaskInput_AssetKey{
		AssetKey: "valid_key",
	}

	permission := &asset.Permission{
		Public:        false,
		AuthorizedIds: []string{owner, defaultWorker},
	}

	function := &asset.Function{
		Key: "c9913ccb-3669-48bb-a6e6-4c84c1cf869a",
		Outputs: map[string]*asset.FunctionOutput{
			"model": {
				Kind: asset.AssetKind_ASSET_MODEL,
			},
			"models": {
				Kind:     asset.AssetKind_ASSET_MODEL,
				Multiple: true,
			},
			"performance": {
				Kind: asset.AssetKind_ASSET_PERFORMANCE,
			},
		},
	}

	validTask := &asset.ComputeTask{
		Key:         "valid_key",
		FunctionKey: function.Key,
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {
				Permissions: &asset.Permissions{
					Download: permission,
					Process:  permission,
				},
			},
			"models": {
				Permissions: &asset.Permissions{
					Download: permission,
					Process:  permission,
				},
			},
		},
	}

	cases := []struct {
		name               string
		worker             string
		function           map[string]*asset.FunctionInput
		task               []*asset.ComputeTaskInput
		dependenciesErrors dependenciesErrors
		expectedError      string
		functionFetched    bool
	}{
		{
			name: "ok",
			function: map[string]*asset.FunctionInput{
				"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
				"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
				"model":       {Kind: asset.AssetKind_ASSET_MODEL},
				"other model": {Kind: asset.AssetKind_ASSET_MODEL},
			},
			task: []*asset.ComputeTaskInput{
				{Identifier: "opener", Ref: validRef},
				{Identifier: "datasamples", Ref: validRef},
				{Identifier: "datasamples", Ref: validRef},
				{Identifier: "model", Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
					ParentTaskKey:    validTask.Key,
					OutputIdentifier: "model",
				}}},
				{Identifier: "other model", Ref: validRef},
			},
			functionFetched: true,
		},
		{
			name:            "optional input",
			function:        map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL, Optional: true}},
			task:            []*asset.ComputeTaskInput{},
			functionFetched: false,
		},
		{
			name:            "missing input",
			function:        map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task:            []*asset.ComputeTaskInput{},
			expectedError:   "missing task input",
			functionFetched: false,
		},
		{
			name:     "duplicate input",
			function: map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task: []*asset.ComputeTaskInput{
				{Identifier: "model", Ref: validRef},
				{Identifier: "model", Ref: validRef},
			},
			expectedError:   "duplicate task input",
			functionFetched: false,
		},
		{
			name:     "unknown input",
			function: map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL, Optional: true}},
			task: []*asset.ComputeTaskInput{
				{Identifier: "foo", Ref: validRef},
			},
			expectedError:   "unknown task input",
			functionFetched: false,
		},
		{
			name:     "error in GetCheckedModel",
			function: map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "model",
					Ref:        validRef,
				},
			},
			dependenciesErrors: dependenciesErrors{
				getCheckedModel: errors.New("model error, e.g. permission error"),
			},
			expectedError:   "model error",
			functionFetched: false,
		},
		{
			name:     "error in GetComputeTask",
			function: map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "model",
					Ref:        &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{ParentTaskKey: validTask.Key}},
				},
			},
			dependenciesErrors: dependenciesErrors{
				getComputeTask: errors.New("task error, e.g. task not found"),
			},
			expectedError:   "task error",
			functionFetched: false,
		},
		{
			name:     "mismatching asset kinds",
			function: map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "model",
					Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
						ParentTaskKey:    validTask.Key,
						OutputIdentifier: "performance",
					}},
				},
			},
			expectedError:   "mismatching task input asset kinds",
			functionFetched: true,
		},
		{
			name:     "parent task output not found",
			function: map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "model",
					Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
						ParentTaskKey:    validTask.Key,
						OutputIdentifier: "not found",
					}},
				},
			},
			expectedError:   "function output not found",
			functionFetched: true,
		},
		{
			name:     "multiple output used as single input",
			function: map[string]*asset.FunctionInput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "model",
					Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
						ParentTaskKey:    validTask.Key,
						OutputIdentifier: "models",
					}},
				},
			},
			expectedError:   "multiple task output used as single task input",
			functionFetched: true,
		},
		{
			name: "input data manager referenced using parent task output",
			function: map[string]*asset.FunctionInput{
				"datamanager": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
				"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "datamanager",
					Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
						ParentTaskKey:    validTask.Key,
						OutputIdentifier: "datamanager",
					}},
				},
				{
					Identifier: "datasamples",
					Ref:        validRef,
				},
			},
			expectedError:   "openers must be referenced using an asset key",
			functionFetched: false,
		},
		{
			name: "error in GetCheckedDataManager",
			function: map[string]*asset.FunctionInput{
				"datamanager": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
				"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "datamanager",
					Ref:        validRef,
				},
				{
					Identifier: "datasamples",
					Ref:        validRef,
				},
			},
			dependenciesErrors: dependenciesErrors{
				checkDataManager: errors.New("data manager error, e.g. permission error"),
			},
			expectedError:   "data manager error",
			functionFetched: false,
		},
		{
			name: "input data sample referenced using parent task output",
			function: map[string]*asset.FunctionInput{
				"datamanager": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
				"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "datamanager",
					Ref:        validRef,
				},
				{
					Identifier: "datasamples",
					Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
						ParentTaskKey:    validTask.Key,
						OutputIdentifier: "datasample",
					}},
				},
			},
			expectedError:   "data samples must be referenced using an asset key",
			functionFetched: false,
		},
		{
			name: "worker is not authorized to process parent task output",
			function: map[string]*asset.FunctionInput{
				"model": {Kind: asset.AssetKind_ASSET_MODEL},
			},
			task: []*asset.ComputeTaskInput{
				{
					Identifier: "model",
					Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
						ParentTaskKey:    validTask.Key,
						OutputIdentifier: "model",
					}},
				},
			},
			expectedError:   "doesn't have permission",
			worker:          "org3",
			functionFetched: true,
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {

				provider := newMockedProvider()
				service := NewComputeTaskService(provider)

				as := new(MockFunctionAPI)
				ctdbal := new(persistence.MockComputeTaskDBAL)
				ms := new(MockModelAPI)
				dms := new(MockDataManagerAPI)

				if c.dependenciesErrors.getComputeTask == nil {
					ctdbal.On("GetComputeTask", mock.Anything).Return(validTask, nil)
				} else {
					ctdbal.On("GetComputeTask", mock.Anything).Return(nil, c.dependenciesErrors.getComputeTask)
				}

				if c.dependenciesErrors.getCheckedModel == nil {
					ms.On("GetCheckedModel", mock.Anything, mock.Anything).Return(&asset.Model{}, nil)
				} else {
					ms.On("GetCheckedModel", mock.Anything, mock.Anything).Return(nil, c.dependenciesErrors.getCheckedModel)
				}
				if c.functionFetched {
					as.On("GetFunction", function.Key).Once().Return(function, nil)
				}

				dataManager := &asset.DataManager{}
				dms.On("GetDataManager", mock.Anything).Once().Return(dataManager, nil)
				dms.On("CheckDataManager", dataManager, mock.Anything, mock.Anything).Return(c.dependenciesErrors.checkDataManager)

				provider.On("GetDataManagerService").Return(dms)
				provider.On("GetModelService").Return(ms)
				provider.On("GetComputeTaskDBAL").Return(ctdbal)
				provider.On("GetFunctionService").Return(as)
				provider.On("GetPermissionService").Return(NewPermissionService(provider))

				worker := defaultWorker
				if c.worker != "" {
					worker = c.worker
				}

				err := service.validateInputs(c.task, c.function, owner, worker)
				if c.expectedError == "" {
					assert.NoError(t, err)
				} else {
					assert.ErrorContains(t, err, c.expectedError)
				}
				as.AssertExpectations(t)
			},
		)
	}
}

func TestValidateTaskOutputs(t *testing.T) {
	cases := []struct {
		name          string
		function      map[string]*asset.FunctionOutput
		task          map[string]*asset.ComputeTaskOutput
		expectedError string
	}{
		{
			name: "ok",
			function: map[string]*asset.FunctionOutput{
				"model": {Kind: asset.AssetKind_ASSET_MODEL},
			},
			task: map[string]*asset.ComputeTaskOutput{
				"model": {},
			},
		},
		{
			name:          "missing output",
			function:      map[string]*asset.FunctionOutput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task:          map[string]*asset.ComputeTaskOutput{},
			expectedError: "missing task output",
		},
		{
			name:     "unknown output",
			function: map[string]*asset.FunctionOutput{"model": {Kind: asset.AssetKind_ASSET_MODEL}},
			task: map[string]*asset.ComputeTaskOutput{
				"foo": {},
			},
			expectedError: "unknown task output",
		},
		{
			name:     "performance permissions",
			function: map[string]*asset.FunctionOutput{"performance": {Kind: asset.AssetKind_ASSET_PERFORMANCE}},
			task: map[string]*asset.ComputeTaskOutput{
				"performance": {Permissions: &asset.Permissions{
					Process:  &asset.Permission{},
					Download: &asset.Permission{},
				}},
			},
			expectedError: "a PERFORMANCE output should be public",
		},
		{
			name:     "performance transient",
			function: map[string]*asset.FunctionOutput{"performance": {Kind: asset.AssetKind_ASSET_PERFORMANCE}},
			task: map[string]*asset.ComputeTaskOutput{
				"performance": {
					Permissions: &asset.Permissions{
						Process:  &asset.Permission{Public: true},
						Download: &asset.Permission{Public: true},
					},
					Transient: true,
				},
			},
			expectedError: "a PERFORMANCE output cannot be transient",
		},
		{
			name:     "performance non transient",
			function: map[string]*asset.FunctionOutput{"performance": {Kind: asset.AssetKind_ASSET_PERFORMANCE}},
			task: map[string]*asset.ComputeTaskOutput{
				"performance": {
					Permissions: &asset.Permissions{
						Process:  &asset.Permission{Public: true},
						Download: &asset.Permission{Public: true},
					},
					Transient: false,
				},
			},
		},
		{
			name:     "public performance",
			function: map[string]*asset.FunctionOutput{"performance": {Kind: asset.AssetKind_ASSET_PERFORMANCE}},
			task: map[string]*asset.ComputeTaskOutput{
				"performance": {Permissions: &asset.Permissions{
					Process:  &asset.Permission{Public: true},
					Download: &asset.Permission{Public: true},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {

				provider := newMockedProvider()
				service := NewComputeTaskService(provider)

				err := service.validateOutputs("uuid", c.task, c.function)
				if c.expectedError == "" {
					assert.NoError(t, err)
				} else {
					assert.ErrorContains(t, err, c.expectedError)
				}
			},
		)
	}
}

func createNode(parent string, key string) *asset.NewComputeTask {
	inputs := []*asset.ComputeTaskInput{}
	if parent != "" {
		inputs = append(inputs, &asset.ComputeTaskInput{
			Identifier: "test",
			Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
				ParentTaskKey:    parent,
				OutputIdentifier: "test",
			}},
		})
	}

	return &asset.NewComputeTask{
		Key:    key,
		Inputs: inputs,
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
		Key: "root",
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
	result, err := service.sortTasks(nodes, existingKeys)

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
		Key: "root",
	}

	leaf1 := createNode(root.Key, "leaf1")
	node2 := createNode(root.Key, "node2")
	node3 := createNode(node2.Key, "node3")
	leaf4 := createNode(node2.Key, "leaf4")
	leaf5 := createNode(node3.Key, "leaf5")

	node2.Inputs = append(node2.Inputs, &asset.ComputeTaskInput{
		Identifier: "test",
		Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
			ParentTaskKey:    leaf4.Key,
			OutputIdentifier: "test",
		}},
	})

	nodes := []*asset.NewComputeTask{root, leaf5, leaf4, node2, node3, leaf1}
	existingKeys := []string{}

	provider := newMockedProvider()
	service := NewComputeTaskService(provider)
	_, err := service.sortTasks(nodes, existingKeys)

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
		Key: "root",
		Inputs: []*asset.ComputeTaskInput{
			{
				Identifier: "test",
				Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
					ParentTaskKey:    existing1,
					OutputIdentifier: "test",
				}},
			},
		},
	}

	leaf1 := createNode(root.Key, "leaf1")
	node2 := createNode(root.Key, "node2")
	node3 := createNode(node2.Key, "node3")
	leaf4 := createNode(node2.Key, "leaf4")
	leaf5 := createNode(node3.Key, "leaf5")

	leaf1.Inputs = append(leaf1.Inputs, &asset.ComputeTaskInput{
		Identifier: "test",
		Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
			ParentTaskKey:    existing2,
			OutputIdentifier: "test",
		}},
	})

	nodes := []*asset.NewComputeTask{root, leaf5, leaf4, node2, node3, leaf1}
	existingKeys := []string{existing1, existing2}

	provider := newMockedProvider()
	service := NewComputeTaskService(provider)
	result, err := service.sortTasks(nodes, existingKeys)

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
		Key: "root",
		Inputs: []*asset.ComputeTaskInput{
			{
				Identifier: "test",
				Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{
					ParentTaskKey: "unknown",
				}},
			},
		},
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
	_, err := service.sortTasks(nodes, existingKeys)

	assert.Error(t, err)
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

func TestDisableOutputs(t *testing.T) {
	t.Run("not worker", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "notmyorg",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "output1", "myorg")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrPermissionDenied, orcError.Kind)

		dbal.AssertExpectations(t)
	})
	t.Run("task not in terminal state", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_EXECUTING,
			Worker: "myorg",
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "output1", "myorg")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrCannotDisableOutput, orcError.Kind)
		assert.Contains(t, orcError.Error(), "task not in final state")

		dbal.AssertExpectations(t)
	})
	t.Run("identifier does not exist", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "myorg",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"output1": {
					Transient: false,
				},
			},
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "nonexistent", "myorg")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrCannotDisableOutput, orcError.Kind)
		assert.Contains(t, orcError.Error(), "there is no output identifier ")

		dbal.AssertExpectations(t)
	})
	t.Run("output is not transient", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "myorg",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"output1": {
					Transient: false,
				},
			},
		}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "output1", "myorg")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrCannotDisableOutput, orcError.Kind)
		assert.Contains(t, orcError.Error(), "output is not transient")

		dbal.AssertExpectations(t)
	})
	t.Run("output kind cannot be deleted", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "myorg",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"output1": {
					Transient: true,
				},
			},
		}

		outputAsset := &asset.ComputeTaskOutputAsset{AssetKind: asset.AssetKind_ASSET_PERFORMANCE}

		dbal := new(persistence.MockDBAL)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		dbal.On("GetComputeTaskOutputAssets", "uuid", "output1").Return([]*asset.ComputeTaskOutputAsset{outputAsset}, nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "output1", "myorg")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrCannotDisableOutput, orcError.Kind)
		assert.Contains(t, orcError.Error(), "cannot disable output of kind")

		dbal.AssertExpectations(t)
	})
	t.Run("children not in final state", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "myorg",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"output1": {
					Transient: true,
				},
			},
		}

		child := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_EXECUTING,
		}

		outputAsset := &asset.ComputeTaskOutputAsset{AssetKind: asset.AssetKind_ASSET_MODEL}

		dbal := new(persistence.MockDBAL)
		modelService := new(MockModelAPI)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		provider.On("GetModelService").Return(modelService)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		dbal.On("GetComputeTaskOutputAssets", "uuid", "output1").Return([]*asset.ComputeTaskOutputAsset{outputAsset}, nil)
		dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{child}, nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "output1", "myorg")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrCannotDisableOutput, orcError.Kind)
		assert.Contains(t, orcError.Error(), "child not in final state")

		dbal.AssertExpectations(t)
		modelService.AssertExpectations(t)
	})
	t.Run("no children", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "myorg",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"output1": {
					Transient: true,
				},
			},
		}

		outputAsset := &asset.ComputeTaskOutputAsset{AssetKind: asset.AssetKind_ASSET_MODEL}

		dbal := new(persistence.MockDBAL)
		modelService := new(MockModelAPI)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		provider.On("GetModelService").Return(modelService)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		dbal.On("GetComputeTaskOutputAssets", "uuid", "output1").Return([]*asset.ComputeTaskOutputAsset{outputAsset}, nil)
		dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{}, nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "output1", "myorg")
		assert.Error(t, err)
		orcError := new(orcerrors.OrcError)
		assert.True(t, errors.As(err, &orcError))
		assert.Equal(t, orcerrors.ErrCannotDisableOutput, orcError.Kind)
		assert.Contains(t, orcError.Error(), "a task with no children")

		dbal.AssertExpectations(t)
		modelService.AssertExpectations(t)
	})
	t.Run("success", func(t *testing.T) {
		task := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "myorg",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"output1": {
					Transient: true,
				},
			},
		}

		child := &asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
		}

		outputAsset := &asset.ComputeTaskOutputAsset{AssetKind: asset.AssetKind_ASSET_MODEL, AssetKey: "modelKey"}

		dbal := new(persistence.MockDBAL)
		modelService := new(MockModelAPI)
		provider := newMockedProvider()
		provider.On("GetComputeTaskDBAL").Return(dbal)
		provider.On("GetModelService").Return(modelService)

		dbal.On("GetComputeTask", "uuid").Return(task, nil)
		dbal.On("GetComputeTaskOutputAssets", "uuid", "output1").Return([]*asset.ComputeTaskOutputAsset{outputAsset}, nil)
		dbal.On("GetComputeTaskChildren", "uuid").Return([]*asset.ComputeTask{child}, nil)

		modelService.On("disable", "modelKey").Return(nil)

		service := NewComputeTaskService(provider)
		err := service.DisableOutput("uuid", "output1", "myorg")
		assert.NoError(t, err)

		dbal.AssertExpectations(t)
		modelService.AssertExpectations(t)
	})
}

func TestRegisterTasksEmptyList(t *testing.T) {
	provider := newMockedProvider()

	service := NewComputeTaskService(provider)

	_, err := service.RegisterTasks([]*asset.NewComputeTask{}, "test")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrBadRequest, orcError.Kind)
}

func TestGetRegisteredTask(t *testing.T) {
	provider := newMockedProvider()
	dbal := new(persistence.MockDBAL)
	provider.On("GetComputeTaskDBAL").Return(dbal)

	service := NewComputeTaskService(provider)

	// simulate tasks in store
	service.taskStore["uuid1"] = &asset.ComputeTask{Key: "uuid1"}
	service.taskStore["uuid3"] = &asset.ComputeTask{Key: "uuid3"}

	// simulate tasks in DB
	dbal.On("GetComputeTasks", []string{"uuid2", "uuid4"}).Once().Return(
		[]*asset.ComputeTask{{Key: "uuid4"}, {Key: "uuid2"}}, // intentionally return them out-of-order because that's what the DB might do
		nil,
	)

	tasks, err := service.getRegisteredTasks("uuid1", "uuid2", "uuid3", "uuid4")
	assert.NoError(t, err)

	assert.Len(t, tasks, 4)

	// The tasks should be returned in the order they were requested
	assert.Equal(t, tasks[0].Key, "uuid1")
	assert.Equal(t, tasks[1].Key, "uuid2")
	assert.Equal(t, tasks[2].Key, "uuid3")
	assert.Equal(t, tasks[3].Key, "uuid4")

	dbal.AssertExpectations(t)
}

func TestGetCachedPlan(t *testing.T) {
	provider := newMockedProvider()
	cps := new(MockComputePlanAPI)
	provider.On("GetComputePlanService").Return(cps)

	computePlan := &asset.ComputePlan{
		Key: "uuid1",
	}

	cps.On("GetPlan", "uuid1").Return(computePlan, nil).Once()

	service := NewComputeTaskService(provider)

	cp, err := service.getCachedCP("uuid1")
	assert.NoError(t, err)
	assert.Equal(t, computePlan.Key, cp.Key)

	cp, err = service.getCachedCP("uuid1")
	assert.NoError(t, err)
	assert.Equal(t, computePlan.Key, cp.Key)

	cps.AssertExpectations(t)
}

func TestGetInputAssetsTaskUnready(t *testing.T) {
	provider := newMockedProvider()
	db := new(persistence.MockComputeTaskDBAL)
	provider.On("GetComputeTaskDBAL").Return(db)

	service := NewComputeTaskService(provider)

	db.On("GetComputeTask", "uuid").
		Once().
		Return(&asset.ComputeTask{
			Key:    "uuid",
			Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS,
		}, nil)

	_, err := service.GetInputAssets("uuid")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "inputs may not be defined")

	provider.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestGetInputAssets(t *testing.T) {
	provider := newMockedProvider()
	as := new(MockFunctionAPI)
	db := new(persistence.MockComputeTaskDBAL)
	dss := new(MockDataSampleAPI)
	dms := new(MockDataManagerAPI)
	ms := new(MockModelAPI)
	provider.On("GetFunctionService").Return(as)
	provider.On("GetComputeTaskDBAL").Return(db)
	provider.On("GetDataSampleService").Return(dss)
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetModelService").Return(ms)

	service := NewComputeTaskService(provider)

	function := &asset.Function{
		Key: "10c97337-e495-4bfd-b189-275b30be8de2",
		Inputs: map[string]*asset.FunctionInput{
			"data":   {Kind: asset.AssetKind_ASSET_DATA_SAMPLE},
			"opener": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"model":  {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}

	as.On("GetFunction", function.Key).Once().Return(function, nil)
	db.On("GetComputeTask", "uuid").
		Once().
		Return(&asset.ComputeTask{
			Key:    "uuid",
			Status: asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT,
			Inputs: []*asset.ComputeTaskInput{
				{Identifier: "data", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:ds"}},
				{Identifier: "opener", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:dm"}},
				{Identifier: "model", Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{ParentTaskKey: "uuid:parent", OutputIdentifier: "aggregate"}}},
			},
			FunctionKey: function.Key,
		}, nil)

	dataSample := &asset.DataSample{Key: "uuid:ds"}
	dataManager := &asset.DataManager{Key: "uuid:dm"}
	model := &asset.Model{Key: "uuid:model"}

	dss.On("GetDataSample", "uuid:ds").
		Once().
		Return(dataSample, nil)

	dms.On("GetDataManager", "uuid:dm").
		Once().
		Return(dataManager, nil)

	db.On("GetComputeTaskOutputAssets", "uuid:parent", "aggregate").
		Once().
		Return(
			[]*asset.ComputeTaskOutputAsset{
				{ComputeTaskKey: "uuid:parent", AssetKind: asset.AssetKind_ASSET_MODEL, AssetKey: "uuid:model"},
			},
			nil,
		)

	ms.On("GetModel", "uuid:model").
		Once().
		Return(model, nil)

	expectedInputs := []*asset.ComputeTaskInputAsset{
		{Identifier: "data", Asset: &asset.ComputeTaskInputAsset_DataSample{DataSample: dataSample}},
		{Identifier: "opener", Asset: &asset.ComputeTaskInputAsset_DataManager{DataManager: dataManager}},
		{Identifier: "model", Asset: &asset.ComputeTaskInputAsset_Model{Model: model}},
	}

	inputAssets, err := service.GetInputAssets("uuid")
	assert.NoError(t, err)

	assert.Equal(t, expectedInputs, inputAssets)

	provider.AssertExpectations(t)
	as.AssertExpectations(t)
	db.AssertExpectations(t)
	dss.AssertExpectations(t)
	dms.AssertExpectations(t)
	ms.AssertExpectations(t)
}

func TestGetParentTaskKeys(t *testing.T) {
	cases := []struct {
		inputs []*asset.ComputeTaskInput
		keys   []string
	}{
		{
			inputs: []*asset.ComputeTaskInput{
				{Identifier: "data", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:ds"}},
				{Identifier: "opener", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:dm"}},
				{Identifier: "model", Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{ParentTaskKey: "uuid:parent", OutputIdentifier: "aggregate"}}},
			},
			keys: []string{"uuid:parent"},
		},
		{
			inputs: []*asset.ComputeTaskInput{
				{Identifier: "local", Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{ParentTaskKey: "uuid:parent", OutputIdentifier: "local"}}},
				{Identifier: "shared", Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{ParentTaskKey: "uuid:parent", OutputIdentifier: "shared"}}},
			},
			keys: []string{"uuid:parent"},
		},
	}

	for i, c := range cases {
		t.Run(
			fmt.Sprintf("parent task keys from inputs case %d", i),
			func(t *testing.T) {
				assert.Equal(t, c.keys, GetParentTaskKeys(c.inputs))
			},
		)
	}
}

func TestGetTaskWorker(t *testing.T) {
	cases := map[string]struct {
		newTask  *asset.NewComputeTask
		function *asset.Function
		err      string
		worker   string
	}{
		"datamanager": {
			newTask: &asset.NewComputeTask{
				Inputs: []*asset.ComputeTaskInput{
					{Identifier: "opener", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:dm1"}},
				},
			},
			function: &asset.Function{
				Inputs: map[string]*asset.FunctionInput{
					"opener": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
				},
			},
			worker: "owner1",
		},
		"worker mismatch": {
			newTask: &asset.NewComputeTask{
				Inputs: []*asset.ComputeTaskInput{
					{Identifier: "opener", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:dm1"}},
				},
				Worker: "worker",
			},
			function: &asset.Function{
				Inputs: map[string]*asset.FunctionInput{
					"opener": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
				},
			},
			err: "OE0003: Specified worker \"worker\" does not match data manager owner: \"owner1\"",
		},
		"aggregation missing worker": {
			newTask: &asset.NewComputeTask{
				Inputs: []*asset.ComputeTaskInput{
					{Identifier: "model", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:model1"}},
					{Identifier: "model", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:model2"}},
				},
			},
			function: &asset.Function{
				Inputs: map[string]*asset.FunctionInput{
					"model": {Kind: asset.AssetKind_ASSET_MODEL},
				},
			},
			err: "OE0003: Worker cannot be inferred and must be explicitly set",
		},
		"aggregation with worker": {
			newTask: &asset.NewComputeTask{
				Inputs: []*asset.ComputeTaskInput{
					{Identifier: "model", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:model1"}},
					{Identifier: "model", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "uuid:model2"}},
				},
				Worker: "worker",
			},
			function: &asset.Function{
				Inputs: map[string]*asset.FunctionInput{
					"model": {Kind: asset.AssetKind_ASSET_MODEL},
				},
			},
			worker: "worker",
		},
	}

	dms := new(MockDataManagerAPI)
	provider := newMockedProvider()
	provider.On("GetDataManagerService").Return(dms)

	dms.On("GetDataManager", "uuid:dm1").Return(&asset.DataManager{Owner: "owner1"}, nil)
	dms.On("GetDataManager", "uuid:dm2").Return(&asset.DataManager{Owner: "owner2"}, nil)

	for name, c := range cases {
		t.Run(
			name,
			func(t *testing.T) {
				service := NewComputeTaskService(provider)

				worker, err := service.getTaskWorker(c.newTask, c.function)
				if c.err != "" {
					assert.Error(t, err)
					assert.EqualError(t, err, c.err)
				}
				assert.Equal(t, c.worker, worker)
			},
		)
	}
}

func TestGetLogsPermission(t *testing.T) {
	dm := &asset.DataManager{
		LogsPermission: &asset.Permission{
			AuthorizedIds: []string{"test"},
			Public:        false,
		},
	}

	dms := new(MockDataManagerAPI)
	os := new(MockOrganizationAPI)
	os.On("GetAllOrganizations").Return([]*asset.Organization{
		{Id: "org1"}, {Id: "org2"}, {Id: "org3"},
	}, nil)

	provider := newMockedProvider()
	provider.On("GetDataManagerService").Return(dms)
	provider.On("GetOrganizationService").Return(os)
	provider.On("GetPermissionService").Return(NewPermissionService(provider))

	dms.On("GetDataManager", "dmuuid").Return(dm, nil)

	cases := map[string]struct {
		owner          string
		parentTasks    []*asset.ComputeTask
		taskInputs     []*asset.ComputeTaskInput
		functionInputs map[string]*asset.FunctionInput
		outcome        *asset.Permission
	}{
		"with datamanager": {
			taskInputs: []*asset.ComputeTaskInput{
				{Identifier: "opener", Ref: &asset.ComputeTaskInput_AssetKey{AssetKey: "dmuuid"}},
			},
			functionInputs: map[string]*asset.FunctionInput{
				"opener": {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			},
			outcome: dm.LogsPermission,
		},
		"with parents": {
			owner: "test",
			taskInputs: []*asset.ComputeTaskInput{
				{Identifier: "model", Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{ParentTaskKey: "parent1", OutputIdentifier: "model"}}},
				{Identifier: "model", Ref: &asset.ComputeTaskInput_ParentTaskOutput{ParentTaskOutput: &asset.ParentTaskOutputRef{ParentTaskKey: "parent2", OutputIdentifier: "model"}}},
			},
			functionInputs: map[string]*asset.FunctionInput{
				"model": {Kind: asset.AssetKind_ASSET_MODEL, Multiple: true},
			},
			parentTasks: []*asset.ComputeTask{
				{Key: "parent1", LogsPermission: &asset.Permission{AuthorizedIds: []string{"org1"}}},
				{Key: "parent2", LogsPermission: &asset.Permission{AuthorizedIds: []string{"org2"}}},
			},
			outcome: &asset.Permission{AuthorizedIds: []string{"test", "org1", "org2"}},
		},
	}

	for name, c := range cases {
		t.Run(
			name,
			func(t *testing.T) {
				service := NewComputeTaskService(provider)

				permission, err := service.getLogsPermission(c.owner, c.parentTasks, c.taskInputs, c.functionInputs)
				assert.NoError(t, err)
				assert.ElementsMatch(t, c.outcome.AuthorizedIds, permission.AuthorizedIds)
				assert.Equal(t, c.outcome.Public, permission.Public)
			},
		)
	}
}
