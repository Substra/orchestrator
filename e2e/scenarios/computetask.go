package scenarios

import (
	"fmt"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var computeTaskTestScenarios = []Scenario{
	{
		testTrainTaskLifecycle,
		[]string{"short", "task"},
	},
	{
		testCascadeCancel,
		[]string{"short", "task"},
	},
	{
		testCascadeTodo,
		[]string{"short", "task"},
	},
	{
		testCascadeFailure,
		[]string{"short", "task"},
	},
	{
		testPropagateLogsPermission,
		[]string{"short", "task", "query"},
	},
	{
		testQueryTasks,
		[]string{"short", "task", "query"},
	},
	{
		testConcurrency,
		[]string{"short", "concurrency"},
	},
	{
		testStableTaskSort,
		[]string{"task", "query"},
	},
	{
		testGetSortedParentTaskKeys,
		[]string{"task", "query"},
	},
}

// Register a task and its dependencies, then start the task.
func testTrainTaskLifecycle(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	ds := appClient.GetDataSample(client.DefaultDataSampleRef)
	if ds.Key != appClient.GetKeyStore().GetKey(client.DefaultDataSampleRef) {
		log.WithField("datasample key", ds.Key).Fatal("datasample could not be properly retrived")
	}

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("anotherTask").WithParentsRef(client.DefaultTaskRef))
	appClient.StartTask(client.DefaultTaskRef)
}

// Register 10 children tasks and cancel their parent
// Only the parent should be canceled
func testCascadeCancel(factory *client.TestClientFactory) {
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
		if task.Status != asset.ComputeTaskStatus_STATUS_WAITING {
			log.Fatal("child task should be WAITING")
		}
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	if plan.Status != asset.ComputePlanStatus_PLAN_STATUS_CANCELED {
		log.WithField("status", plan.Status).Fatal("compute plan has not the CANCELED status")
	}
}

// Register 10 tasks and set their parent as done
func testCascadeTodo(factory *client.TestClientFactory) {
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
		if task.Status != asset.ComputeTaskStatus_STATUS_TODO {
			log.Fatal("child task should be TODO")
		}
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	if plan.Status != asset.ComputePlanStatus_PLAN_STATUS_DOING {
		log.WithField("status", plan.Status).Fatal("compute plan has not the DOING status")
	}
}

// Register 10 tasks and set their parent as failed
// Only the parent should be FAILED
func testCascadeFailure(factory *client.TestClientFactory) {
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
		if task.Status != asset.ComputeTaskStatus_STATUS_WAITING {
			log.Fatal("child task should be WAITING")
		}
	}

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	if plan.Status != asset.ComputePlanStatus_PLAN_STATUS_FAILED {
		log.WithField("status", plan.Status).Fatal("compute plan has not the FAILED status")
	}
}

func testPropagateLogsPermission(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())

	logsPermission := &asset.NewPermissions{Public: false, AuthorizedIds: []string{appClient.MSPID}}
	datamanager := client.DefaultDataManagerOptions().WithLogsPermission(logsPermission)
	appClient.RegisterDataManager(datamanager)

	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	task := appClient.GetComputeTask(client.DefaultTaskRef)
	if !(len(task.LogsPermission.AuthorizedIds) == 1 && task.LogsPermission.AuthorizedIds[0] == appClient.MSPID) {
		log.Fatal("Unexpected task.LogsPermission.")
	}
}

func testQueryTasks(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultAlgoRef)}, "", 10)

	if len(resp.Tasks) != 1 {
		log.WithField("num_tasks", len(resp.Tasks)).Fatal("unexpected task result")
	}
}

func testConcurrency(factory *client.TestClientFactory) {
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

// testStableTaskSort will register several hundreds of tasks and query them all multiple time, failing if there are duplicates.
func testStableTaskSort(factory *client.TestClientFactory) {
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
					log.WithField("iteration", i+1).Fatal("Unstable task sorting: duplicate task found")
				}
				keys[task.Key] = struct{}{}
			}
			resp = getPage(resp.NextPageToken)
		}
	}
}

// testGetSortedParentTaskKeys will check that parent task keys are returned in the same order they were registered
func testGetSortedParentTaskKeys(factory *client.TestClientFactory) {
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

	if len(task.ParentTaskKeys) != len(parentTaskRefs) {
		log.Fatal("Unexpected ParentTaskKeys length")
	}

	for i, taskRef := range parentTaskRefs {
		if task.ParentTaskKeys[i] != appClient.GetKeyStore().GetKey(taskRef) {
			log.Fatal("Unexpected ParentTaskKeys ordering")
		}
	}
}
