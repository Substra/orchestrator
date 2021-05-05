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

// Register a task and its dependencies, then start the task.
func testTrainTaskLifecycle(conn *grpc.ClientConn) {
	defer log.WithTrace().Info("test train task lifecycle")
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-train-task-lifecycle")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan()
	appClient.RegisterTrainTask()
	appClient.RegisterChildTask("anotherTask")
	appClient.StartTrainTask()
}

// register a task, start it, and register a model on it.
func testRegisterModel(conn *grpc.ClientConn) {
	defer log.WithTrace().Info("test register model")
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-register-model")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan()
	appClient.AssertPlanInQueryPlans()
	appClient.RegisterTrainTask()
	appClient.StartTrainTask()
	appClient.RegisterModel()
}

// Register 10 children tasks and cancel their parent
func testCascadeCancel(conn *grpc.ClientConn) {
	defer log.WithTrace().Info("test cascade 10 tasks CANCELED")
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-cancel-tasks")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan()
	appClient.AssertPlanInQueryPlans()
	appClient.RegisterTrainTask()

	for i := 0; i < 10; i++ {
		appClient.RegisterChildTask(fmt.Sprintf("task%d", i))
	}

	appClient.StartTrainTask()
	appClient.CancelTrainTask()

	task := appClient.GetComputeTask("task3")
	if task.Status != asset.ComputeTaskStatus_STATUS_CANCELED {
		log.Fatal("child task should be CANCELED")
	}
}

// Register 10 tasks and set their parent as done
func testCascadeTodo(conn *grpc.ClientConn) {
	defer log.WithTrace().Info("test cascade 10 tasks TODO")
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-todo-tasks")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan()
	appClient.AssertPlanInQueryPlans()
	appClient.RegisterTrainTask()

	for i := 0; i < 10; i++ {
		appClient.RegisterChildTask(fmt.Sprintf("task%d", i))
	}

	appClient.StartTrainTask()
	appClient.DoneTrainTask()

	task := appClient.GetComputeTask("task3")
	if task.Status != asset.ComputeTaskStatus_STATUS_TODO {
		log.Fatal("child task should be TODO")
	}
}

// Register 10 tasks and set their parent as done
func testCascadeFailure(conn *grpc.ClientConn) {
	defer log.WithTrace().Info("test cascade 10 tasks FAILED")
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-fail-tasks")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan()
	appClient.AssertPlanInQueryPlans()
	appClient.RegisterTrainTask()

	for i := 0; i < 10; i++ {
		appClient.RegisterChildTask(fmt.Sprintf("task%d", i))
	}

	appClient.StartTrainTask()
	appClient.FailTrainTask()

	task := appClient.GetComputeTask("task3")
	if task.Status != asset.ComputeTaskStatus_STATUS_CANCELED {
		log.Fatal("child task should be CANCELED")
	}
}
