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
	appClient.RegisterModel(client.DefaultTaskRef, "model")
}

// Register 10 children tasks and cancel their parent
func testCascadeCancel(conn *grpc.ClientConn) {
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
	appClient.RegisterModel(client.DefaultTaskRef, "model0")
	appClient.DoneTask(client.DefaultTaskRef)
	// second done
	appClient.StartTask("child1")
	appClient.RegisterModel("child1", "model1")
	appClient.DoneTask("child1")
	// last task
	appClient.StartTask("child2")
	appClient.RegisterModel("child2", "model2")
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
		client.DefaultAggregateTaskOptions().WithKeyRef("compC4").WithParentsRef([]string{"compA3", "compB3"}).WithAlgoRef("algoAgg"),
	)

	lastAggregate := appClient.GetComputeTask("compC4")
	if lastAggregate.Rank != 3 {
		log.WithField("rank", lastAggregate.Rank).Fatal("last aggegation task has not expected rank")
	}
}
