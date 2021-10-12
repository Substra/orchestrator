package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/grpc"
)

var defaultParent = []string{client.DefaultTaskRef}

type scenario struct {
	exec func(*grpc.ClientConn)
	tags []string
}

var testScenarios = map[string]scenario{
	"TrainTaskLifecycle": {
		testTrainTaskLifecycle,
		[]string{"short", "task"},
	},
	"RegisterModel": {
		testRegisterModel,
		[]string{"short", "model"},
	},
	"CascadeCancel": {
		testCascadeCancel,
		[]string{"short", "task"},
	},
	"CascadeTodo": {
		testCascadeTodo,
		[]string{"short", "task"},
	},
	"CascadeFailure": {
		testCascadeFailure,
		[]string{"short", "task"},
	},
	"DeleteIntermediary": {
		testDeleteIntermediary,
		[]string{"short", "model"},
	},
	"MultiStageComputePlan": {
		testMultiStageComputePlan,
		[]string{"short", "plan"},
	},
	"QueryTasks": {
		testQueryTasks,
		[]string{"short", "task", "query"},
	},
	"RegisterPerformance": {
		testRegisterPerformance,
		[]string{"short", "perf"},
	},
	"CompositeParentChild": {
		testCompositeParentChild,
		[]string{"short", "plan"},
	},
	"Concurrency": {
		testConcurrency,
		[]string{"short", "concurrency"},
	},
	"LargeComputePlan": {
		testLargeComputePlan,
		[]string{"long", "plan"},
	},
	"BatchLargeComputePlan": {
		testBatchLargeComputePlan,
		[]string{"long", "plan"},
	},
	"SmallComputePlan": {
		testSmallComputePlan,
		[]string{"short", "plan"},
	},
	"AggregateComposite": {
		testAggregateComposite,
		[]string{"short", "plan", "aggregate", "composite"},
	},
	"DatasetSampleKeys": {
		testDatasetSampleKeys,
		[]string{"short", "dataset"},
	},
	"QueryAlgos": {
		testQueryAlgos,
		[]string{"short", "algo"},
	},
	"FailLargeComputePlan": {
		testFailLargeComputePlan,
		[]string{"long", "plan"},
	},
	"StableTaskSort": {
		testStableTaskSort,
		[]string{"task", "query"},
	},
	"QueryComputePlan": {
		testQueryComputePlan,
		[]string{"short", "plan", "query"},
	},
}

// Register a task and its dependencies, then start the task.
func testTrainTaskLifecycle(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("anotherTask").WithParentsRef(defaultParent...))
	appClient.StartTask(client.DefaultTaskRef)
}

// register a task, start it, and register a model on it.
func testRegisterModel(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	plan := appClient.GetComputePlan("cp")
	if plan.TaskCount != 1 {
		log.Fatal("compute plan has invalid task count")
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

	taskEvents := appClient.QueryEvents(&asset.EventQueryFilter{AssetKey: appClient.GetKeyStore().GetKey(client.DefaultTaskRef)}, "", 10)

	if len(taskEvents.Events) != 3 {
		// 3 events: creation, start, done
		log.WithField("events", taskEvents.Events).Fatal("Unexpected number of events")
	}
}

// Register 10 children tasks and cancel their parent
// Only the parent should be canceled
func testCascadeCancel(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent...))
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
func testCascadeTodo(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent...))
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
func testCascadeFailure(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent...))
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

// register 3 successive tasks, start and register models then check for model deletion
func testDeleteIntermediary(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithDeleteIntermediaryModels(true))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child1").WithParentsRef(defaultParent...))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child2").WithParentsRef("child1"))

	// First task done
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model0"))
	// second done
	appClient.StartTask("child1")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model1").WithTaskRef("child1"))
	// last task
	appClient.StartTask("child2")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model2").WithTaskRef("child2"))

	models := appClient.GetTaskOutputModels(client.DefaultTaskRef)
	if len(models) != 1 {
		log.Fatal("invalid number of output models")
	}

	if models[0].Address == nil {
		log.Fatal("model has a invalid address")
	}

	if !appClient.CanDisableModel("model0") {
		log.Fatal("parent model cannot be disabled")
	}

	if appClient.CanDisableModel("model2") {
		log.Fatal("final model can be disabled")
	}

	appClient.DisableModel("model0")
	models = appClient.GetTaskOutputModels(client.DefaultTaskRef)
	if models[0].Address != nil {
		log.Fatal("model has not been disabled")
	}
}

// This is the "canonical" example of FL with 2 nodes aggregating their trunks
// This does not check multi-organization setup though!
//
//   ,========,                ,========,
//   | ORG A  |                | ORG B  |
//   *========*                *========*
//
//     ø     ø                  ø      ø
//     |     |                  |      |
//     hd    tr                 tr     hd
//   -----------              -----------
//  | Composite |            | Composite |      STEP 1
//   -----------              -----------
//     hd    tr                 tr     hd
//     |      \   ,========,   /      |
//     |       \  | ORG C  |  /       |
//     |        \ *========* /        |
//     |       ----------------       |
//     |      |    Aggregate   |      |         STEP 2
//     |       ----------------       |
//     |              |               |
//     |     ,_______/ \_______       |
//     |     |                 |      |
//     hd    tr                tr     hd
//   -----------             -----------
//  | Composite |           | Composite |       STEP 3
//   -----------             -----------
//     hd    tr                 tr     hd
//            \                /
//             \              /
//              \            /
//             ----------------
//            |    Aggregate   |                STEP 4
//             ----------------
//
//
func testMultiStageComputePlan(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("algoComp").WithCategory(asset.AlgoCategory_ALGO_COMPOSITE))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("algoAgg").WithCategory(asset.AlgoCategory_ALGO_AGGREGATE))
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	// step 1
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compA1").WithAlgoRef("algoComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compB1").WithAlgoRef("algoComp"),
	)
	// step 2
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().WithKeyRef("aggC2").WithParentsRef("compA1", "compB1").WithAlgoRef("algoAgg"),
	)
	// step 3
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compA3").WithParentsRef("compA1", "aggC2").WithAlgoRef("algoComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compB3").WithParentsRef("compB1", "aggC2").WithAlgoRef("algoComp"),
	)
	// step 4
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().WithKeyRef("aggC4").WithParentsRef("compA3", "compB3").WithAlgoRef("algoAgg"),
	)

	lastAggregate := appClient.GetComputeTask("aggC4")
	if lastAggregate.Rank != 3 {
		log.WithField("rank", lastAggregate.Rank).Fatal("last aggegation task has not expected rank")
	}

	// Start step 1
	appClient.StartTask("compA1")
	appClient.StartTask("compB1")

	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA1").WithKeyRef("modelA1H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA1").WithKeyRef("modelA1T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB1").WithKeyRef("modelB1H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB1").WithKeyRef("modelB1T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	cp := appClient.GetComputePlan(client.DefaultPlanRef)
	if cp.Status != asset.ComputePlanStatus_PLAN_STATUS_DOING {
		log.WithField("status", cp.Status).Fatal("unexpected compute plan status")
	}
	if cp.DoneCount != 2 {
		log.WithField("doneCount", cp.DoneCount).Fatal("invalid task count")
	}

	// Start step 2
	appClient.StartTask("aggC2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("aggC2").WithKeyRef("modelC2").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	// Start step 3
	appClient.StartTask("compA3")
	appClient.StartTask("compB3")

	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA3").WithKeyRef("modelA3H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA3").WithKeyRef("modelA3T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB3").WithKeyRef("modelB3H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB3").WithKeyRef("modelB3T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	// Start step 4
	appClient.StartTask("aggC4")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("aggC4").WithKeyRef("modelC4").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
}

func testQueryTasks(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultAlgoRef)}, "", 10)

	if len(resp.Tasks) != 1 {
		log.WithField("num_tasks", len(resp.Tasks)).Fatal("unexpected task result")
	}
}

// register a test task, start it, and register its performance.
func testRegisterPerformance(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterMetric(client.DefaultMetricOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	// Parent train task is necessary
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	// to create a test task
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithDataSampleRef("testds").WithParentsRef(defaultParent...))
	appClient.StartTask("testTask")
	appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask"))

	appClient.GetTaskPerformance("testTask")

	task := appClient.GetComputeTask("testTask")
	if task.Status != asset.ComputeTaskStatus_STATUS_DONE {
		log.Fatal("test task should be DONE")
	}
}

func testCompositeParentChild(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("algoComp").WithCategory(asset.AlgoCategory_ALGO_COMPOSITE))
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("comp1").WithAlgoRef("algoComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("comp2").WithAlgoRef("algoComp").WithParentsRef("comp1", "comp1"),
	)

	appClient.StartTask("comp1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	appClient.StartTask("comp2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
}

func testConcurrency(conn *grpc.ClientConn) {
	client1, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create client1")
	}
	client2, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create client2")
	}
	// Share the same key store for both clients
	client2.WithKeyStore(client1.GetKeyStore())

	client1.EnsureNode()
	client1.RegisterAlgo(client.DefaultAlgoOptions())
	client1.RegisterDataManager()
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

func testLargeComputePlan(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	nbTasks := 10000
	nbQuery := 5000 // 10k exceed max response size

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	start := time.Now()
	for i := 0; i < nbTasks; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent...))
	}
	log.WithField("registrationDuration", time.Since(start)).WithField("nbTasks", nbTasks).Info("registration done")

	start = time.Now()
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultAlgoRef)}, "", nbQuery)
	log.WithField("queryDuration", time.Since(start)).WithField("nbTasks", nbQuery).Info("query done")

	if len(resp.Tasks) != nbQuery {
		log.WithField("nbTasks", len(resp.Tasks)).Fatal("unexpected task count")
	}
}

func testBatchLargeComputePlan(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	nbTasks := 10000
	batchSize := 1000
	nbQuery := 5000 // 10k exceed max response size

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	start := time.Now()
	for i := 0; i < nbTasks; {
		batchStart := time.Now()
		newTasks := make([]client.Taskable, 0, batchSize)
		for c := 0; c < batchSize && i < nbTasks; c++ {
			i++
			newTasks = append(newTasks, client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent...))
		}
		appClient.RegisterTasks(newTasks...)
		log.WithField("batchDuration", time.Since(batchStart)).WithField("nbTasks", i).Info("batch done")
	}
	log.WithField("registrationDuration", time.Since(start)).WithField("nbTasks", nbTasks).Info("registration done")

	start = time.Now()
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultAlgoRef)}, "", nbQuery)
	log.WithField("queryDuration", time.Since(start)).WithField("nbTasks", nbQuery).Info("query done")

	if len(resp.Tasks) != nbQuery {
		log.WithField("nbTasks", len(resp.Tasks)).Fatal("unexpected task count")
	}
}

func testSmallComputePlan(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterMetric(client.DefaultMetricOptions())

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),
		client.DefaultTrainTaskOptions().WithKeyRef("train3").WithParentsRef("train1", "train2"),
		client.DefaultTestTaskOptions().WithDataSampleRef("objSample").WithParentsRef("train3"),
	)

	cp := appClient.GetComputePlan(client.DefaultPlanRef)
	if cp.Status != asset.ComputePlanStatus_PLAN_STATUS_TODO {
		log.WithField("status", cp.Status).Fatal("unexpected plan status")
	}
	if cp.TaskCount != 4 {
		log.WithField("taskCount", cp.TaskCount).Fatal("invalid task count")
	}
}

func testAggregateComposite(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_COMPOSITE))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_AGGREGATE).WithKeyRef("aggAlgo"))
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterMetric(client.DefaultMetricOptions())

	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("c1"),
		client.DefaultCompositeTaskOptions().WithKeyRef("c2"),
		client.DefaultAggregateTaskOptions().WithKeyRef("a1").WithAlgoRef("aggAlgo").WithParentsRef("c1", "c2"),
		client.DefaultCompositeTaskOptions().WithKeyRef("c3").WithParentsRef("a1", "c1"),
	)

	appClient.StartTask("c1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c1").WithKeyRef("m1H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c1").WithKeyRef("m1T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	appClient.StartTask("c2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c2").WithKeyRef("m2H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c2").WithKeyRef("m2T").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	appClient.StartTask("a1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("a1").WithKeyRef("mAgg").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	appClient.StartTask("c3")

	inputs := appClient.GetInputModels("c3")
	if len(inputs) != 2 {
		log.Fatal("composite should have 2 input models")
	}
}

func testDatasetSampleKeys(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.EnsureNode()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds2"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithTestOnly(true).WithKeyRef("testds"))

	dataset := appClient.GetDataset(client.DefaultDataManagerRef)

	if len(dataset.TestDataSampleKeys) != 1 {
		log.Fatal("dataset should contain a single test sample")
	}
	if len(dataset.TrainDataSampleKeys) != 2 {
		log.Fatal("dataset should contain 2 train samples")
	}
	if dataset.TestDataSampleKeys[0] != appClient.GetKeyStore().GetKey("testds") {
		log.Fatal("dataset should contain valid test sample ID")
	}
}

func testQueryAlgos(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterMetric(client.DefaultMetricOptions())

	resp := appClient.QueryAlgos(&asset.AlgoQueryFilter{}, "", 100)

	// We cannot check for equality since this test may run after others,
	// we will probably have more than the registered algo above.
	if len(resp.Algos) < 1 {
		log.WithField("numAlgos", len(resp.Algos)).Fatal("Unexpected total number of algo")
	}

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	planKey := appClient.GetKeyStore().GetKey(client.DefaultPlanRef)

	resp = appClient.QueryAlgos(&asset.AlgoQueryFilter{ComputePlanKey: planKey}, "", 100)
	if len(resp.Algos) != 0 {
		log.WithField("numAlgos", len(resp.Algos)).Fatal("Unexpected number algo used in compute plan without tasks")
	}

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),
		client.DefaultTrainTaskOptions().WithKeyRef("train3").WithParentsRef("train1", "train2"),
		client.DefaultTestTaskOptions().WithDataSampleRef("objSample").WithParentsRef("train3"),
	)

	resp = appClient.QueryAlgos(&asset.AlgoQueryFilter{ComputePlanKey: planKey}, "", 100)
	if len(resp.Algos) != 1 {
		log.WithField("numAlgos", len(resp.Algos)).Fatal("Unexpected number of algo used in compute plan with tasks")
	}
}

func testFailLargeComputePlan(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	nbRounds := 1000
	nbPharma := 11
	var nbTasks int

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("algoComp").WithCategory(asset.AlgoCategory_ALGO_COMPOSITE))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("algoAgg").WithCategory(asset.AlgoCategory_ALGO_AGGREGATE))
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	newTasks := make([]client.Taskable, 0)
	start := time.Now()
	for i := 0; i < nbRounds; {
		compKeys := make([]string, nbPharma)

		for pharma := 1; pharma < nbPharma+1; {
			compKey := fmt.Sprintf("compP%dR%d", pharma, i)
			compKeys[pharma-1] = compKey

			task := client.DefaultCompositeTaskOptions().WithKeyRef(compKey).WithAlgoRef("algoComp")
			if i > 0 {
				// Reference previous composite and aggregate
				task.WithParentsRef(fmt.Sprintf("compP%dR%d", pharma, i-1), fmt.Sprintf("aggR%d", i-1))
			}
			newTasks = append(newTasks, task)
			nbTasks++
			pharma++
		}

		// Add aggregate
		newTasks = append(newTasks, client.DefaultAggregateTaskOptions().WithKeyRef(fmt.Sprintf("aggR%d", i)).WithParentsRef(compKeys...).WithAlgoRef("algoAgg"))
		nbTasks++

		i++

		if i%20 == 0 {
			appClient.RegisterTasks(newTasks...)
			log.WithField("round", i).WithField("nbTasks", nbTasks).WithField("duration", time.Since(start)).Debug("Round registered")
			newTasks = make([]client.Taskable, 0)
			start = time.Now()
		}
	}

	if len(newTasks) > 0 {
		appClient.RegisterTasks(newTasks...)
		log.WithField("nbTasks", nbTasks).WithField("duration", time.Since(start)).Debug("Round registered")
	}

	// Fail the composite of rank 0 on pharma1
	start = time.Now()
	appClient.StartTask("compP1R0")
	appClient.FailTask("compP1R0")
	log.WithField("duration", time.Since(start)).WithField("nbTasks", nbTasks).Info("canceled compute plan")
}

// testStableTaskSort will register several hundreds of tasks and query them all multiple time, failing if there are duplicates.
func testStableTaskSort(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	nbTasks := 1000
	nbQuery := 10

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	newTasks := make([]client.Taskable, 0, nbTasks)
	for i := 0; i < nbTasks; i++ {
		newTasks = append(newTasks, client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent...))
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

func testQueryComputePlan(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef("cp1"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef("cp2"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef("cp3"))

	resp := appClient.QueryPlans("", 3)

	if len(resp.Plans) != 3 {
		log.WithField("nbPlans", len(resp.Plans)).Fatal("Unexpected number of compute plans")
	}
}
