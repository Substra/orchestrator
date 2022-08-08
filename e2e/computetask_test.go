//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"sync"
	"testing"

	"github.com/owkin/orchestrator/e2e/client"
	e2erequire "github.com/owkin/orchestrator/e2e/require"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterComputeTask(t *testing.T) {
	appClient := factory.NewTestClient()
	ks := appClient.GetKeyStore()
	computePlanRef := "register task cp"

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef(computePlanRef))
	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
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
	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	ds := appClient.GetDataSample(client.DefaultDataSampleRef)
	require.Equal(t, appClient.GetKeyStore().GetKey(client.DefaultDataSampleRef), ds.Key, "datasample could not be properly retrieved")

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("anotherTask").
		WithParentsRef(client.DefaultTrainTaskRef).
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask(client.DefaultTrainTaskRef)
}

func TestPredictTaskLifecycle(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions().WithKeyRef("train_algo"))
	appClient.RegisterAlgo(client.DefaultPredictAlgoOptions().WithKeyRef("predict_algo"))
	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions().WithKeyRef("metric"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("train").WithAlgoRef("train_algo"))

	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithParentsRef("train").
		WithInput("model", &client.TaskOutputRef{TaskRef: "train", Identifier: "model"}).
		WithAlgoRef("predict_algo").
		WithKeyRef("predict"))

	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("test").
		WithParentsRef("predict").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predict", Identifier: "predictions"}).
		WithAlgoRef("metric"))

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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).
			WithParentsRef(client.DefaultTrainTaskRef).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}

	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.CancelTask(client.DefaultTrainTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		require.Equal(t, asset.ComputeTaskStatus_STATUS_WAITING, task.Status, "child task should be WAITING")
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_CANCELED, plan.Status)
}

// TestCascadeTodo registers 10 tasks and set their parent as done
func TestCascadeTodo(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).
			WithParentsRef(client.DefaultTrainTaskRef).
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
	require.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_DOING, plan.Status)
}

// TestCascadeFailure registers 10 tasks and set their parent as failed
// Only the parent should be FAILED
func TestCascadeFailure(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).
			WithParentsRef(client.DefaultTrainTaskRef).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}

	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.FailTask(client.DefaultTrainTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		require.Equal(t, asset.ComputeTaskStatus_STATUS_WAITING, task.Status, "child task should be WAITING")
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_FAILED, plan.Status)
}

func TestPropagateLogsPermission(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())

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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultSimpleAlgoRef)}, "", 10)
	require.Equal(t, 1, len(resp.Tasks))
}

func TestConcurrency(t *testing.T) {
	client1 := factory.NewTestClient()
	client2 := factory.NewTestClient()

	// Share the same key store for both clients
	client2.WithKeyStore(client1.GetKeyStore())

	client1.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	client1.RegisterDataManager(client.DefaultDataManagerOptions())
	client1.RegisterDataSample(client.DefaultDataSampleOptions())
	client1.RegisterComputePlan(client.DefaultComputePlanOptions())

	client1.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("parent1"))
	client2.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("parent2"))

	wg := new(sync.WaitGroup)

	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func(i int) {
			client1.RegisterTasks(client.DefaultTrainTaskOptions().
				WithKeyRef(fmt.Sprintf("task1%d", i)).
				WithParentsRef("parent1").
				WithInput("model", &client.TaskOutputRef{TaskRef: "parent1", Identifier: "model"}))
			wg.Done()
		}(i)
		go func(i int) {
			client2.RegisterTasks(client.DefaultTrainTaskOptions().
				WithKeyRef(fmt.Sprintf("task2%d", i)).
				WithParentsRef("parent2").
				WithInput("model", &client.TaskOutputRef{TaskRef: "parent2", Identifier: "model"}))
			wg.Done()
		}(i)
	}

	wg.Wait()
}

// TestStableTaskSort will register several hundreds of tasks and query them all multiple time, failing if there are duplicates.
func TestStableTaskSort(t *testing.T) {
	appClient := factory.NewTestClient()

	nbTasks := 1000
	nbQuery := 10

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	newTasks := make([]client.Taskable, 0, nbTasks)
	for i := 0; i < nbTasks; i++ {
		newTasks = append(newTasks, client.DefaultTrainTaskOptions().
			WithKeyRef(fmt.Sprintf("task%d", i)).
			WithParentsRef(client.DefaultTrainTaskRef).
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

// TestGetSortedParentTaskKeys will check that parent task keys are returned in the same order they were registered
func TestGetSortedParentTaskKeys(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions().WithKeyRef("trainAlgo"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	const nbParents int = 10
	parentTaskRefs := make([]string, nbParents)
	for i := 0; i < nbParents; i++ {
		parentTaskRefs[i] = fmt.Sprint("parent", i)
		appClient.RegisterTasks(
			client.DefaultTrainTaskOptions().WithKeyRef(parentTaskRefs[i]).WithAlgoRef("trainAlgo"),
		)
	}

	appClient.RegisterAlgo(client.DefaultAggregateAlgoOptions().WithKeyRef("aggAlgo"))
	agg := client.
		DefaultAggregateTaskOptions().
		WithAlgoRef("aggAlgo").
		WithKeyRef("aggTask").
		WithParentsRef(parentTaskRefs...)
	for _, parent := range parentTaskRefs {
		agg.WithInput("model", &client.TaskOutputRef{TaskRef: parent, Identifier: "model"})
	}
	appClient.RegisterTasks(agg)

	task := appClient.GetComputeTask("aggTask")
	require.Equal(t, len(parentTaskRefs), len(task.ParentTaskKeys))

	for i, taskRef := range parentTaskRefs {
		require.Equal(t, task.ParentTaskKeys[i], appClient.GetKeyStore().GetKey(taskRef), "unexpected ParentTaskKeys ordering")
	}
}

func TestQueryTaskInputs(t *testing.T) {
	appClient := factory.NewTestClient()
	ks := appClient.GetKeyStore()

	taskRef := "task with inputs"
	cpRef := "CP with inputs"
	parentAlgoRef := "parent algo"
	childAlgoRef := "child algo"
	parentTaskRef := "parent task"
	otherTaskRef := "other task"
	inputModelRef := "input model"

	parentAlgoOptions := client.
		DefaultSimpleAlgoOptions().
		WithKeyRef(parentAlgoRef).
		WithOutput("output model", asset.AssetKind_ASSET_MODEL, false)

	childAlgoOptions := client.
		DefaultSimpleAlgoOptions().
		WithKeyRef(childAlgoRef).
		WithOutput("other model", asset.AssetKind_ASSET_MODEL, true)

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterAlgo(parentAlgoOptions)
	appClient.RegisterAlgo(childAlgoOptions)
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef(cpRef))

	parentTaskOptions := client.
		DefaultTrainTaskOptions().
		WithKeyRef(parentTaskRef).
		WithAlgoRef(parentAlgoRef).
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
		WithAlgoRef(childAlgoRef).
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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
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

	client1.RegisterAlgo(client.DefaultSimpleAlgoOptions().WithKeyRef("trainAlgo"))
	client1.RegisterDataManager(client.DefaultDataManagerOptions())
	client1.RegisterDataSample(client.DefaultDataSampleOptions())
	client1.RegisterComputePlan(client.DefaultComputePlanOptions())

	client1.RegisterTasks(
		client.DefaultTrainTaskOptions().WithAlgoRef("trainAlgo").WithKeyRef("trainTask1"),
		client.DefaultTrainTaskOptions().WithAlgoRef("trainAlgo").WithKeyRef("trainTask2"),
	)

	client1.RegisterAlgo(client.DefaultAggregateAlgoOptions().WithKeyRef("aggAlgo"))
	client1.RegisterTasks(client.DefaultAggregateTaskOptions().
		WithAlgoRef("aggAlgo").
		WithKeyRef("aggTask").
		WithParentsRef("trainTask1").
		WithInput("model", &client.TaskOutputRef{TaskRef: "trainTask1", Identifier: "model"}).
		WithWorker("MyOrg2MSP"))

	client1.StartTask("trainTask1")
	client1.RegisterModel(client.DefaultModelOptions().WithTaskRef("trainTask1"))

	client1.StartTask("trainTask2")
	client1.FailTask("trainTask2")
	client2.CancelTask("aggTask")
}
