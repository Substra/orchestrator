//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	// "sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
)

func TestRegisterComputeTask(t *testing.T) {
	appClient := factory.NewTestClient()
	ks := appClient.GetKeyStore()
	computePlanRef := "register task cp"

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef(computePlanRef))
	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	registeredTask := appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithPlanRef(computePlanRef))[0]

	// GetTask
	retrievedTask := appClient.GetComputeTask(client.DefaultTrainTaskRef)
	e2erequire.ProtoEqual(t, registeredTask, retrievedTask)

	// QueryTasks
	retrievedTask = appClient.QueryTasks(&asset.TaskQueryFilter{ComputePlanKey: ks.GetKey(computePlanRef)}, "", 10).Tasks[0]
	e2erequire.ProtoEqual(t, registeredTask, retrievedTask)
}

func TestRegisterTaskWithTransientOutput(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	registeredTask := appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithOutput("model", &asset.NewPermissions{Public: true}, true))[0]

	// GetTask
	retrievedTask := appClient.GetComputeTask(client.DefaultTrainTaskRef)
	e2erequire.ProtoEqual(t, registeredTask, retrievedTask)
	assert.True(t, retrievedTask.Outputs["model"].Transient)
}

// TestTrainTaskLifecycle registers a task and its dependencies, then start the task.
func TestTrainTaskLifecycle(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	ds := appClient.GetDataSample(client.DefaultDataSampleRef)
	require.Equal(t, appClient.GetKeyStore().GetKey(client.DefaultDataSampleRef), ds.Key, "datasample could not be properly retrieved")

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("anotherTask").
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask(client.DefaultTrainTaskRef)
}

func TestPredictTaskLifecycle(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions().WithKeyRef("train_function"))
	appClient.RegisterFunction(client.DefaultPredictFunctionOptions().WithKeyRef("predict_function"))
	appClient.RegisterFunction(client.DefaultMetricFunctionOptions().WithKeyRef("metric"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("train").WithFunctionRef("train_function"))

	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithInput("model", &client.TaskOutputRef{TaskRef: "train", Identifier: "model"}).
		WithFunctionRef("predict_function").
		WithKeyRef("predict"))

	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("test").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predict", Identifier: "predictions"}).
		WithFunctionRef("metric"))

	appClient.StartTask("train")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("train").WithKeyRef("train_end"))
	appClient.DoneTask("train")

	appClient.StartTask("predict")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predict").WithKeyRef("pred_end").WithTaskOutput("predictions"))
	appClient.DoneTask("predict")

	predictTask := appClient.GetComputeTask("predict")
	require.Equal(t, predictTask.Status, asset.ComputeTaskStatus_STATUS_DONE)

	testTask := appClient.GetComputeTask("test")
	require.Equal(t, testTask.Status, asset.ComputeTaskStatus_STATUS_TODO)
}

// TestCascadeCancel registers 10 children tasks and cancel their parent
// Only the parent should be canceled
func TestCascadeCancel(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}

	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.CancelTask(client.DefaultTrainTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		require.Equal(t, asset.ComputeTaskStatus_STATUS_WAITING, task.Status, "child task should be WAITING")
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Nil(t, plan.CancelationDate)
	require.Nil(t, plan.FailureDate)
}

// TestCascadeTodo registers 10 tasks and set their parent as done
func TestCascadeTodo(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}

	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		require.Equal(t, asset.ComputeTaskStatus_STATUS_TODO, task.Status, "child task should be TODO")
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Nil(t, plan.CancelationDate)
	require.Nil(t, plan.FailureDate)
}

// TestCascadeFailure registers 10 tasks and set their parent as failed
// Only the parent should be FAILED
func TestCascadeFailure(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}

	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.FailTask(client.DefaultTrainTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		require.Equal(t, asset.ComputeTaskStatus_STATUS_WAITING, task.Status, "child task should be WAITING")
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.NotNil(t, plan.FailureDate)
	require.Nil(t, plan.CancelationDate)
}

func TestPropagateLogsPermission(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())

	logsPermission := &asset.NewPermissions{Public: false, AuthorizedIds: []string{appClient.MSPID}}
	datamanager := client.DefaultDataManagerOptions().WithLogsPermission(logsPermission)
	appClient.RegisterDataManager(datamanager)

	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	task := appClient.GetComputeTask(client.DefaultTrainTaskRef)
	require.Equal(t, 1, len(task.LogsPermission.AuthorizedIds))
	require.Equal(t, appClient.MSPID, task.LogsPermission.AuthorizedIds[0])
}

func TestQueryTasks(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	resp := appClient.QueryTasks(&asset.TaskQueryFilter{FunctionKey: appClient.GetKeyStore().GetKey(client.DefaultSimpleFunctionRef)}, "", 10)
	require.Equal(t, 1, len(resp.Tasks))
}

// func TestConcurrency(t *testing.T) {
// 	client1 := factory.NewTestClient()
// 	client2 := factory.NewTestClient()

// 	// Share the same key store for both clients
// 	client2.WithKeyStore(client1.GetKeyStore())

// 	client1.RegisterFunction(client.DefaultSimpleFunctionOptions())
// 	client1.RegisterDataManager(client.DefaultDataManagerOptions())
// 	client1.RegisterDataSample(client.DefaultDataSampleOptions())
// 	client1.RegisterComputePlan(client.DefaultComputePlanOptions())

// 	client1.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("parent1"))
// 	client2.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("parent2"))

// 	wg := new(sync.WaitGroup)

// 	for i := 0; i < 5; i++ {
// 		wg.Add(2)
// 		go func(i int) {
// 			client1.RegisterTasks(client.DefaultTrainTaskOptions().
// 				WithKeyRef(fmt.Sprintf("task1%d", i)).
// 				WithInput("model", &client.TaskOutputRef{TaskRef: "parent1", Identifier: "model"}))
// 			wg.Done()
// 		}(i)
// 		go func(i int) {
// 			client2.RegisterTasks(client.DefaultTrainTaskOptions().
// 				WithKeyRef(fmt.Sprintf("task2%d", i)).
// 				WithInput("model", &client.TaskOutputRef{TaskRef: "parent2", Identifier: "model"}))
// 			wg.Done()
// 		}(i)
// 	}

// 	wg.Wait()
// }

// TestStableTaskSort will register several hundreds of tasks and query them all multiple time, failing if there are duplicates.
func TestStableTaskSort(t *testing.T) {
	appClient := factory.NewTestClient()

	nbTasks := 1000
	nbQuery := 10

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	newTasks := make([]client.Taskable, 0, nbTasks)
	for i := 0; i < nbTasks; i++ {
		newTasks = append(newTasks, client.DefaultTrainTaskOptions().
			WithKeyRef(fmt.Sprintf("task%d", i)).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}
	appClient.RegisterTasks(newTasks...)

	getPage := func(token string) *asset.QueryTasksResponse {
		return appClient.QueryTasks(
			&asset.TaskQueryFilter{ComputePlanKey: appClient.GetKeyStore().GetKey(client.DefaultPlanRef)},
			token,
			100,
		)
	}

	for i := 0; i < nbQuery; i++ {
		keys := make(map[string]struct{})
		resp := getPage("")

		for resp.NextPageToken != "" {
			for _, task := range resp.Tasks {
				if _, ok := keys[task.Key]; ok {
					require.FailNow(t, fmt.Sprintf("unstable task sorting: duplicate task found (iteration %d)", i+1))
				}
				keys[task.Key] = struct{}{}
			}
			resp = getPage(resp.NextPageToken)
		}
	}
}

func TestQueryTaskInputs(t *testing.T) {
	appClient := factory.NewTestClient()
	ks := appClient.GetKeyStore()

	taskRef := "task with inputs"
	cpRef := "CP with inputs"
	parentFunctionRef := "parent function"
	childFunctionRef := "child function"
	parentTaskRef := "parent task"
	otherTaskRef := "other task"
	inputModelRef := "input model"

	parentFunctionOptions := client.
		DefaultSimpleFunctionOptions().
		WithKeyRef(parentFunctionRef).
		WithOutput("output model", asset.AssetKind_ASSET_MODEL, false)

	childFunctionOptions := client.
		DefaultSimpleFunctionOptions().
		WithKeyRef(childFunctionRef).
		WithOutput("other model", asset.AssetKind_ASSET_MODEL, true)

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterFunction(parentFunctionOptions)
	appClient.RegisterFunction(childFunctionOptions)
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef(cpRef))

	parentTaskOptions := client.
		DefaultTrainTaskOptions().
		WithKeyRef(parentTaskRef).
		WithFunctionRef(parentFunctionRef).
		WithPlanRef(cpRef).
		WithOutput("output model", &asset.NewPermissions{
			Public:        false,
			AuthorizedIds: []string{appClient.MSPID},
		}, false)
	appClient.RegisterTasks(parentTaskOptions)

	otherTaskOptions := client.DefaultTrainTaskOptions().WithKeyRef(otherTaskRef).WithPlanRef(cpRef)
	appClient.RegisterTasks(otherTaskOptions)
	appClient.StartTask(otherTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef(inputModelRef).WithTaskRef(otherTaskRef))

	taskOptions := client.
		DefaultTrainTaskOptions().
		WithFunctionRef(childFunctionRef).
		WithKeyRef(taskRef).
		WithPlanRef(cpRef).
		WithInputAsset("model", inputModelRef).
		WithInput("model", &client.TaskOutputRef{TaskRef: parentTaskRef, Identifier: "output model"}).
		WithOutput("other model", &asset.NewPermissions{
			Public:        false,
			AuthorizedIds: []string{appClient.MSPID},
		}, false)

	expectedInputs := []*asset.ComputeTaskInput{
		{
			Identifier: "opener",
			Ref: &asset.ComputeTaskInput_AssetKey{
				AssetKey: ks.GetKey(client.DefaultDataManagerRef),
			},
		},
		{
			Identifier: "datasamples",
			Ref: &asset.ComputeTaskInput_AssetKey{
				AssetKey: ks.GetKey(client.DefaultDataSampleRef),
			},
		},
		{
			Identifier: "model",
			Ref: &asset.ComputeTaskInput_AssetKey{
				AssetKey: ks.GetKey(inputModelRef),
			},
		},
		{
			Identifier: "model",
			Ref: &asset.ComputeTaskInput_ParentTaskOutput{
				ParentTaskOutput: &asset.ParentTaskOutputRef{
					ParentTaskKey:    ks.GetKey(parentTaskRef),
					OutputIdentifier: "output model",
				},
			},
		},
	}

	expectedOutputs := map[string]*asset.ComputeTaskOutput{
		"model": {
			Permissions: &asset.Permissions{
				Process: &asset.Permission{
					Public:        true, // yes, we have public = true + non empty AuthorizedIds. Blame newPermission()'s behavior. :shrug:
					AuthorizedIds: []string{appClient.MSPID},
				},
				Download: &asset.Permission{
					Public:        true,
					AuthorizedIds: []string{appClient.MSPID},
				},
			},
		},
		"other model": {
			Permissions: &asset.Permissions{
				Process: &asset.Permission{
					Public:        false,
					AuthorizedIds: []string{appClient.MSPID},
				},
				Download: &asset.Permission{
					Public:        false,
					AuthorizedIds: []string{appClient.MSPID},
				},
			},
		},
	}

	// test RegisterTasks
	tasks := appClient.RegisterTasks(taskOptions)
	task := tasks[0]
	e2erequire.ProtoArrayEqual(t, task.Inputs, expectedInputs)
	e2erequire.ProtoMapEqual(t, task.Outputs, expectedOutputs)

	// test GetComputeTask
	respTask := appClient.GetComputeTask(taskRef)
	e2erequire.ProtoArrayEqual(t, respTask.Inputs, expectedInputs)
	e2erequire.ProtoMapEqual(t, respTask.Outputs, expectedOutputs)

	// test QueryTasks
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{ComputePlanKey: ks.GetKey(cpRef)}, "", 10000)
	found := false
	for _, task := range resp.Tasks {
		if task.Key == ks.GetKey(taskRef) {
			found = true
			e2erequire.ProtoArrayEqual(t, task.Inputs, expectedInputs)
			e2erequire.ProtoMapEqual(t, task.Outputs, expectedOutputs)
			break
		}
	}

	require.True(t, found, "Could not find expected task with key ref "+taskRef)
}

func TestEventsDuringComputeTaskLifecycle(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	registeredTask := appClient.RegisterTasks(client.DefaultTrainTaskOptions())[0]

	getEventTask := func(eventKind asset.EventKind) *asset.ComputeTask {
		res := appClient.QueryEvents(&asset.EventQueryFilter{
			AssetKey:  registeredTask.Key,
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: eventKind,
		}, "", 100)

		require.Len(t, res.Events, 1)
		return res.Events[0].GetComputeTask()
	}

	registrationEventTask := getEventTask(asset.EventKind_EVENT_ASSET_CREATED)
	e2erequire.ProtoEqual(t, registeredTask, registrationEventTask)

	appClient.StartTask(client.DefaultTrainTaskRef)
	startedTask := appClient.GetComputeTask(client.DefaultTrainTaskRef)

	startEventTask := getEventTask(asset.EventKind_EVENT_ASSET_UPDATED)
	e2erequire.ProtoEqual(t, startedTask, startEventTask)
}

// TestWorkerCancelTaskInFailedComputePlan ensures that a worker can cancel a task it does not own
// following the failure of the compute plan.
func TestWorkerCancelTaskInFailedComputePlan(t *testing.T) {
	client1 := factory.WithMSPID("MyOrg1MSP").NewTestClient()
	client2 := factory.WithMSPID("MyOrg2MSP").NewTestClient().WithKeyStore(client1.GetKeyStore())

	client1.RegisterFunction(client.DefaultSimpleFunctionOptions().WithKeyRef("trainFunction"))
	client1.RegisterDataManager(client.DefaultDataManagerOptions())
	client1.RegisterDataSample(client.DefaultDataSampleOptions())
	client1.RegisterComputePlan(client.DefaultComputePlanOptions())

	client1.RegisterTasks(
		client.DefaultTrainTaskOptions().WithFunctionRef("trainFunction").WithKeyRef("trainTask1"),
		client.DefaultTrainTaskOptions().WithFunctionRef("trainFunction").WithKeyRef("trainTask2"),
	)

	client1.RegisterFunction(client.DefaultAggregateFunctionOptions().WithKeyRef("aggFunction"))
	client1.RegisterTasks(client.DefaultAggregateTaskOptions().
		WithFunctionRef("aggFunction").
		WithKeyRef("aggTask").
		WithInput("model", &client.TaskOutputRef{TaskRef: "trainTask1", Identifier: "model"}).
		WithWorker("MyOrg2MSP"))

	client1.StartTask("trainTask1")
	client1.RegisterModel(client.DefaultModelOptions().WithTaskRef("trainTask1"))

	client1.StartTask("trainTask2")
	client1.FailTask("trainTask2")
	client2.CancelTask("aggTask")
}

// TestGetTaskInputAssets creates two trains and an aggregate tasks
// and assert that their inputs are those expected.
func TestGetTaskInputAssets(t *testing.T) {
	appClient := factory.NewTestClient()
	ks := appClient.GetKeyStore()

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions().WithKeyRef("train_function"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train").WithFunctionRef("train_function"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2").WithFunctionRef("train_function"), // Make sure we have several models as aggregate input to check order
	)

	// Check that train has expected inputs
	trainInputs := appClient.GetTaskInputAssets("train")
	assert.Len(t, trainInputs, 2)

	for _, input := range trainInputs {
		if input.Identifier == "opener" {
			assert.IsType(t, &asset.ComputeTaskInputAsset_DataManager{}, input.Asset)
			dm := input.Asset.(*asset.ComputeTaskInputAsset_DataManager).DataManager
			assert.Equal(t, ks.GetKey(client.DefaultDataManagerRef), dm.Key)
		}
		if input.Identifier == "datasamples" {
			assert.IsType(t, &asset.ComputeTaskInputAsset_DataSample{}, input.Asset)
			ds := input.Asset.(*asset.ComputeTaskInputAsset_DataSample).DataSample
			assert.Equal(t, ks.GetKey(client.DefaultDataSampleRef), ds.Key)
		}
	}

	appClient.RegisterFunction(
		client.DefaultAggregateFunctionOptions().
			SetInputs(map[string]*asset.FunctionInput{
				"sink": {Kind: asset.AssetKind_ASSET_MODEL, Multiple: true},
			}).
			WithKeyRef("aggregate_function"),
	)
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().
			WithInput("sink", &client.TaskOutputRef{TaskRef: "train", Identifier: "model"}).
			WithInput("sink", &client.TaskOutputRef{TaskRef: "train2", Identifier: "model"}).
			WithKeyRef("aggregate").
			WithFunctionRef("aggregate_function"),
	)

	// Fetching inputs for a task not in TODO should return an error
	_, err := appClient.FailableGetTaskInputAssets("aggregate")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "inputs may not be defined")

	appClient.StartTask("train")
	appClient.RegisterModels(client.DefaultModelOptions().WithTaskRef("train").WithKeyRef("train_model"))
	appClient.DoneTask("train")

	appClient.StartTask("train2")
	appClient.RegisterModels(client.DefaultModelOptions().WithTaskRef("train2").WithKeyRef("train2_model"))
	appClient.DoneTask("train2")

	aggInputs := appClient.GetTaskInputAssets("aggregate")
	assert.Len(t, aggInputs, 2)
	expectedKeyRef := []string{"train_model", "train2_model"}
	for i, input := range aggInputs {
		assert.IsType(t, &asset.ComputeTaskInputAsset_Model{}, input.Asset)
		model := input.Asset.(*asset.ComputeTaskInputAsset_Model).Model
		keyRef := expectedKeyRef[i]
		assert.Equal(t, ks.GetKey(keyRef), model.Key)
	}
}

func TestGetTaskInputAssetsFromComposite(t *testing.T) {
	appClient := factory.NewTestClient()
	ks := appClient.GetKeyStore()

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterFunction(client.DefaultCompositeFunctionOptions().WithKeyRef("comp_function"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("comp1").WithFunctionRef("comp_function"),
		client.DefaultCompositeTaskOptions().WithKeyRef("comp2").WithFunctionRef("comp_function"),
	)

	// Check that composite has expected inputs
	compositeInputs := appClient.GetTaskInputAssets("comp1")
	assert.Len(t, compositeInputs, 2)

	for _, input := range compositeInputs {
		if input.Identifier == "opener" {
			assert.IsType(t, &asset.ComputeTaskInputAsset_DataManager{}, input.Asset)
			dm := input.Asset.(*asset.ComputeTaskInputAsset_DataManager).DataManager
			assert.Equal(t, ks.GetKey(client.DefaultDataManagerRef), dm.Key)
		}
		if input.Identifier == "datasamples" {
			assert.IsType(t, &asset.ComputeTaskInputAsset_DataSample{}, input.Asset)
			ds := input.Asset.(*asset.ComputeTaskInputAsset_DataSample).DataSample
			assert.Equal(t, ks.GetKey(client.DefaultDataSampleRef), ds.Key)
		}
	}

	appClient.RegisterFunction(
		client.DefaultAggregateFunctionOptions().
			SetInputs(map[string]*asset.FunctionInput{
				"sink": {Kind: asset.AssetKind_ASSET_MODEL, Multiple: true},
			}).
			WithKeyRef("aggregate_function"),
	)
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().
			WithInput("sink", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "shared"}).
			WithInput("sink", &client.TaskOutputRef{TaskRef: "comp2", Identifier: "shared"}).
			WithKeyRef("aggregate").
			WithFunctionRef("aggregate_function"),
	)

	// Fetching inputs for a task not in TODO should return an error
	_, err := appClient.FailableGetTaskInputAssets("aggregate")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "inputs may not be defined")

	appClient.StartTask("comp1")
	appClient.RegisterModels(
		client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("local1").WithTaskOutput("local"),
		client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("shared1").WithTaskOutput("shared"),
	)
	appClient.DoneTask("comp1")

	appClient.StartTask("comp2")
	// Here we register local after shared: order should not matter
	appClient.RegisterModels(
		client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("shared2").WithTaskOutput("shared"),
		client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("local2").WithTaskOutput("local"),
	)
	appClient.DoneTask("comp2")

	aggInputs := appClient.GetTaskInputAssets("aggregate")
	assert.Len(t, aggInputs, 2)
	expectedKeyRef := []string{"shared1", "shared2"}
	for i, input := range aggInputs {
		assert.IsType(t, &asset.ComputeTaskInputAsset_Model{}, input.Asset)
		model := input.Asset.(*asset.ComputeTaskInputAsset_Model).Model
		keyRef := expectedKeyRef[i]
		assert.Equal(t, ks.GetKey(keyRef), model.Key)
	}
}
