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

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/grpc"
)

var defaultParent = []string{client.DefaultTaskRef}

// Register a task and its dependencies, then start the task.
func testTrainTaskLifecycle(conn *grpc.ClientConn) {
	log.Debug("testTrainTaskLifecycle")
	defer log.WithTrace().Info("test train task lifecycle")

	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions())
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions().WithKeyRef("anotherTask").WithParentsRef(defaultParent))
	appClient.StartTask(client.DefaultTaskRef)
}

// register a task, start it, and register a model on it.
func testRegisterModel(conn *grpc.ClientConn) {
	log.Debug("testRegisterModel")
	defer log.WithTrace().Info("test register model")

	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions())

	plan := appClient.GetComputePlan("cp")
	if plan.TaskCount != 1 {
		log.Fatal("compute plan has invalid task count")
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
}

// Register 10 children tasks and cancel their parent
func testCascadeCancel(conn *grpc.ClientConn) {
	log.Debug("testCascadeCancel")
	defer log.WithTrace().Info("test cascade 10 tasks CANCELED")

	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTrainTask(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.CancelTask(client.DefaultTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		if task.Status != asset.ComputeTaskStatus_STATUS_CANCELED {
			log.Fatal("child task should be CANCELED")
		}
	}
}

// Register 10 tasks and set their parent as done
func testCascadeTodo(conn *grpc.ClientConn) {
	log.Debug("testCascadeTodo")
	defer log.WithTrace().Info("test cascade 10 tasks TODO")

	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTrainTask(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.DoneTask(client.DefaultTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		if task.Status != asset.ComputeTaskStatus_STATUS_TODO {
			log.Fatal("child task should be TODO")
		}
	}
}

// Register 10 tasks and set their parent as failed
func testCascadeFailure(conn *grpc.ClientConn) {
	log.Debug("testCascadeFailure")
	defer log.WithTrace().Info("test cascade 10 tasks FAILED")

	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTrainTask(client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTask(client.DefaultTaskRef)
	appClient.FailTask(client.DefaultTaskRef)

	for i := 0; i < 10; i++ {
		task := appClient.GetComputeTask(fmt.Sprintf("task%d", i))
		if task.Status != asset.ComputeTaskStatus_STATUS_CANCELED {
			log.Fatal("child task should be CANCELED")
		}
	}

}

// register 3 successive tasks, start and register models then check for model deletion
func testDeleteIntermediary(conn *grpc.ClientConn) {
	log.Debug("testDeleteIntermediary")
	defer log.WithTrace().Info("test disabling intermediary models")

	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithDeleteIntermediaryModels(true))
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions())

	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions().WithKeyRef("child1").WithParentsRef(defaultParent))
	appClient.RegisterTrainTask(client.DefaultTrainTaskOptions().WithKeyRef("child2").WithParentsRef([]string{"child1"}))

	// First task done
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model0"))
	appClient.DoneTask(client.DefaultTaskRef)
	// second done
	appClient.StartTask("child1")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model1").WithTaskRef("child1"))
	appClient.DoneTask("child1")
	// last task
	appClient.StartTask("child2")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model2").WithTaskRef("child2"))
	appClient.DoneTask("child2")

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
	log.Debug("testMultiStageComputePlan")
	defer log.WithTrace().Info("test multi stage compute plan")

	appClient, err := client.NewTestClient(conn, *mspid, *channel, *chaincode)
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.EnsureNode()
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("algoComp").WithCategory(asset.AlgoCategory_ALGO_COMPOSITE))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef("algoAgg").WithCategory(asset.AlgoCategory_ALGO_AGGREGATE))
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	// step 1
	appClient.RegisterCompositeTask(
		client.DefaultCompositeTaskOptions().WithKeyRef("compA1").WithAlgoRef("algoComp"),
	)
	appClient.RegisterCompositeTask(
		client.DefaultCompositeTaskOptions().WithKeyRef("compB1").WithAlgoRef("algoComp"),
	)
	// step 2
	appClient.RegisterAggregateTask(
		client.DefaultAggregateTaskOptions().WithKeyRef("aggC2").WithParentsRef([]string{"compA1", "compB1"}).WithAlgoRef("algoAgg"),
	)
	// step 3
	appClient.RegisterCompositeTask(
		client.DefaultCompositeTaskOptions().WithKeyRef("compA3").WithParentsRef([]string{"compA1", "aggC2"}).WithAlgoRef("algoComp"),
	)
	appClient.RegisterCompositeTask(
		client.DefaultCompositeTaskOptions().WithKeyRef("compB3").WithParentsRef([]string{"compB1", "aggC2"}).WithAlgoRef("algoComp"),
	)
	// step 4
	appClient.RegisterAggregateTask(
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

	appClient.DoneTask("compA1")
	appClient.DoneTask("compB1")

	// Start step 2
	appClient.StartTask("aggC2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("aggC2").WithKeyRef("modelC2").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
	appClient.DoneTask("aggC2")

	// Start step 3
	appClient.StartTask("compA3")
	appClient.StartTask("compB3")

	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA3").WithKeyRef("modelA3H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compA3").WithKeyRef("modelA3T").WithCategory(asset.ModelCategory_MODEL_TRUNK))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB3").WithKeyRef("modelB3H").WithCategory(asset.ModelCategory_MODEL_HEAD))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("compB3").WithKeyRef("modelB3T").WithCategory(asset.ModelCategory_MODEL_TRUNK))

	appClient.DoneTask("compA3")
	appClient.DoneTask("compB3")

	// Start step 4
	appClient.StartTask("aggC4")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("aggC4").WithKeyRef("modelC4").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
	appClient.DoneTask("aggC4")
}
