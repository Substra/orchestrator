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
		testSmallCp,
		[]string{"short", "plan"},
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
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("anotherTask").WithParentsRef(defaultParent))
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
}

// Register 10 children tasks and cancel their parent
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
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.CancelTask(client.DefaultTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		if task.Status != asset.ComputeTaskStatus_STATUS_CANCELED {
			log.Fatal("child task should be CANCELED")
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
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
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
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.FailTask(client.DefaultTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		if task.Status != asset.ComputeTaskStatus_STATUS_CANCELED {
			log.Fatal("child task should be CANCELED")
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

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child1").WithParentsRef(defaultParent))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child2").WithParentsRef([]string{"child1"}))

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
		client.DefaultAggregateTaskOptions().WithKeyRef("aggC2").WithParentsRef([]string{"compA1", "compB1"}).WithAlgoRef("algoAgg"),
	)
	// step 3
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compA3").WithParentsRef([]string{"compA1", "aggC2"}).WithAlgoRef("algoComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compB3").WithParentsRef([]string{"compB1", "aggC2"}).WithAlgoRef("algoComp"),
	)
	// step 4
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().WithKeyRef("aggC4").WithParentsRef([]string{"compA3", "compB3"}).WithAlgoRef("algoAgg"),
	)

	lastAggregate := appClient.GetComputeTask("aggC4")
	if lastAggregate.Rank != 3 {
		log.WithField("rank", lastAggregate.Rank).Fatal("last aggegation task has not expected rank")
	}

	// Start step 1
	appClient.StartTask("compA1")
	appClient.StartTask("compB1")

	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA1").WithKeyRef("modelA1H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA1").WithKeyRef("modelA1T").WithCategory(asset.ModelCategory_MODEL_TRUNK))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB1").WithKeyRef("modelB1H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB1").WithKeyRef("modelB1T").WithCategory(asset.ModelCategory_MODEL_TRUNK))

	// Start step 2
	appClient.StartTask("aggC2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("aggC2").WithKeyRef("modelC2").WithCategory(asset.ModelCategory_MODEL_SIMPLE))

	// Start step 3
	appClient.StartTask("compA3")
	appClient.StartTask("compB3")

	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA3").WithKeyRef("modelA3H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA3").WithKeyRef("modelA3T").WithCategory(asset.ModelCategory_MODEL_TRUNK))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB3").WithKeyRef("modelB3H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB3").WithKeyRef("modelB3T").WithCategory(asset.ModelCategory_MODEL_TRUNK))

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
	appClient.RegisterObjective(client.DefaultObjectiveOptions().WithDataSampleRef("testds"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	// Parent train task is necessary
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	// to create a test task
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithParentsRef(defaultParent))
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
		client.DefaultCompositeTaskOptions().WithKeyRef("comp2").WithAlgoRef("algoComp").WithParentsRef([]string{"comp1", "comp1"}),
	)

	appClient.StartTask("comp1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1T").WithCategory(asset.ModelCategory_MODEL_TRUNK))

	appClient.StartTask("comp2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2T").WithCategory(asset.ModelCategory_MODEL_TRUNK))
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
			client1.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task1%d", i)).WithParentsRef([]string{"parent1"}))
			wg.Done()
		}(i)
		go func(i int) {
			client2.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task2%d", i)).WithParentsRef([]string{"parent2"}))
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
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
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
			newTasks = append(newTasks, client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
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

func testSmallCp(conn *grpc.ClientConn) {
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
	appClient.RegisterObjective(client.DefaultObjectiveOptions().WithDataSampleRef("objSample"))

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),
		client.DefaultTrainTaskOptions().WithKeyRef("train3").WithParentsRef([]string{"train1", "train2"}),
		client.DefaultTestTaskOptions().WithParentsRef([]string{"train3"}),
	)
}
