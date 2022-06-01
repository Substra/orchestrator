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
	"github.com/stretchr/testify/require"
)

func TestRegisterComputeTask(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	registeredTask := appClient.RegisterTasks(client.DefaultTrainTaskOptions())[0]
	retrievedTask := appClient.GetComputeTask(client.DefaultTaskRef)

	e2erequire.ProtoEqual(t, registeredTask, retrievedTask)
}

// TestTrainTaskLifecycle registers a task and its dependencies, then start the task.
func TestTrainTaskLifecycle(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	ds := appClient.GetDataSample(client.DefaultDataSampleRef)
	require.Equal(t, appClient.GetKeyStore().GetKey(client.DefaultDataSampleRef), ds.Key, "datasample could not be properly retrieved")

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("anotherTask").WithParentsRef(client.DefaultTaskRef))
	appClient.StartTask(client.DefaultTaskRef)
}

func TestPredictTaskLifecycle(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_PREDICT).WithKeyRef("predict_algo"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_SIMPLE).WithKeyRef("train_algo"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("metric"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("train").WithAlgoRef("train_algo"))
	appClient.RegisterTasks(client.DefaultPredictTaskOptions().WithParentsRef("train").WithAlgoRef("predict_algo").WithKeyRef("predict"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("test").WithParentsRef("predict").WithAlgoRef("train_algo").WithMetricsRef("metric"))

	appClient.StartTask("train")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("train").WithKeyRef("train_end"))

	appClient.StartTask("predict")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predict").WithKeyRef("pred_end"))

	predictTask := appClient.GetComputeTask("predict")
	require.Equal(t, predictTask.Status, asset.ComputeTaskStatus_STATUS_DONE)

	testTask := appClient.GetComputeTask("test")
	require.Equal(t, testTask.Status, asset.ComputeTaskStatus_STATUS_TODO)
}

// TestCascadeCancel registers 10 children tasks and cancel their parent
// Only the parent should be canceled
func TestCascadeCancel(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(client.DefaultTaskRef))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.CancelTask(client.DefaultTaskRef)

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

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(client.DefaultTaskRef))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

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

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(client.DefaultTaskRef))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.FailTask(client.DefaultTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		require.Equal(t, asset.ComputeTaskStatus_STATUS_WAITING, task.Status, "child task should be WAITING")
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_FAILED, plan.Status)
}

func TestPropagateLogsPermission(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())

	logsPermission := &asset.NewPermissions{Public: false, AuthorizedIds: []string{appClient.MSPID}}
	datamanager := client.DefaultDataManagerOptions().WithLogsPermission(logsPermission)
	appClient.RegisterDataManager(datamanager)

	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	task := appClient.GetComputeTask(client.DefaultTaskRef)
	require.Equal(t, 1, len(task.LogsPermission.AuthorizedIds))
	require.Equal(t, appClient.MSPID, task.LogsPermission.AuthorizedIds[0])
}

func TestQueryTasks(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultAlgoRef)}, "", 10)
	require.Equal(t, 1, len(resp.Tasks))
}

func TestConcurrency(t *testing.T) {
	client1 := factory.NewTestClient()
	client2 := factory.NewTestClient()

	// Share the same key store for both clients
	client2.WithKeyStore(client1.GetKeyStore())

	client1.RegisterAlgo(client.DefaultAlgoOptions())
	client1.RegisterDataManager(client.DefaultDataManagerOptions())
	client1.RegisterDataSample(client.DefaultDataSampleOptions())
	client1.RegisterComputePlan(client.DefaultComputePlanOptions())

	client1.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("parent1"))
	client2.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("parent2"))

	wg := new(sync.WaitGroup)

	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func(i int) {
			client1.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task1%d", i)).WithParentsRef("parent1"))
			wg.Done()
		}(i)
		go func(i int) {
			client2.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task2%d", i)).WithParentsRef("parent2"))
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

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	newTasks := make([]client.Taskable, 0, nbTasks)
	for i := 0; i < nbTasks; i++ {
		newTasks = append(newTasks, client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(client.DefaultTaskRef))
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

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("trainAlgo"))
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

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_AGGREGATE).WithKeyRef("aggAlgo"))
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().
			WithAlgoRef("aggAlgo").
			WithKeyRef("aggTask").
			WithParentsRef(parentTaskRefs...),
	)

	task := appClient.GetComputeTask("aggTask")
	require.Equal(t, len(parentTaskRefs), len(task.ParentTaskKeys))

	for i, taskRef := range parentTaskRefs {
		require.Equal(t, task.ParentTaskKeys[i], appClient.GetKeyStore().GetKey(taskRef), "unexpected ParentTaskKeys ordering")
	}
}

func TestQueryTaskInputs(t *testing.T) {
	appClient := factory.NewTestClient()
	ks := appClient.GetKeyStore()

	taskRef := "taskWithInputs"
	cpRef := "taskWithInputs"

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef(cpRef))

	parentTaskOptions := client.
		DefaultTrainTaskOptions().
		WithKeyRef("parentTaskRef").
		WithPlanRef(cpRef)

	appClient.RegisterTasks(parentTaskOptions)

	taskOptions := client.
		DefaultTrainTaskOptions().
		WithKeyRef(taskRef).
		WithPlanRef(cpRef).
		WithAssetInputRef("models", "inputModelRef").
		WithParentTaskInputRef("models", "parentTaskRef", "output model")

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
			Identifier: "models",
			Ref: &asset.ComputeTaskInput_AssetKey{
				AssetKey: ks.GetKey("inputModelRef"),
			},
		},
		{
			Identifier: "models",
			Ref: &asset.ComputeTaskInput_ParentTaskOutput{
				ParentTaskOutput: &asset.ParentTaskOutputRef{
					ParentTaskKey:    ks.GetKey("parentTaskRef"),
					OutputIdentifier: "output model",
				},
			},
		},
	}

	// test RegisterTasks
	tasks := appClient.RegisterTasks(taskOptions)
	task := tasks[0]
	e2erequire.ProtoArrayEqual(t, task.Inputs, expectedInputs)

	// test GetComputeTask
	respTask := appClient.GetComputeTask(taskRef)
	e2erequire.ProtoArrayEqual(t, respTask.Inputs, expectedInputs)

	// test QueryTasks
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{ComputePlanKey: ks.GetKey(cpRef)}, "", 2)
	found := false
	for _, task := range resp.Tasks {
		if task.Key == ks.GetKey(taskRef) {
			found = true
			e2erequire.ProtoArrayEqual(t, task.Inputs, expectedInputs)
			break
		}
	}

	require.True(t, found, "Could not find expected task with key ref "+taskRef)
}

func TestEventsDuringComputeTaskLifecycle(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
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

	appClient.StartTask(client.DefaultTaskRef)
	startedTask := appClient.GetComputeTask(client.DefaultTaskRef)

	startEventTask := getEventTask(asset.EventKind_EVENT_ASSET_UPDATED)
	e2erequire.ProtoEqual(t, startedTask, startEventTask)
}
