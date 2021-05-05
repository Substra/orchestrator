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
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-train-task-lifecycle")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(nil)
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions())
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions().WithKeyRef("anotherTask").WithParentsRef(defaultParent))
	appClient.StartTrainTask(client.DefaultTaskRef)
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
	appClient.RegisterComputePlan(nil)
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions())

	plan := appClient.GetComputePlan("cp")
	if plan.TaskCount != 1 {
		log.Fatal("compute plan has invalid task count")
	}

	appClient.StartTrainTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultTaskRef, "model")
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
	appClient.RegisterComputePlan(nil)
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTrainTask(client.DefaultTaskRef)
	appClient.CancelTrainTask(client.DefaultTaskRef)

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
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-todo-tasks")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(nil)
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTrainTask(client.DefaultTaskRef)
	appClient.DoneTrainTask(client.DefaultTaskRef)

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
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-fail-tasks")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(nil)
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions())

	for i := 0; i < 10; i++ {
		appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(defaultParent))
	}

	appClient.StartTrainTask(client.DefaultTaskRef)
	appClient.FailTrainTask(client.DefaultTaskRef)

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
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-delete-intermediary")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan(&client.ComputePlanOptions{DeleteIntermediaryModels: true})
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions())

	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions().WithKeyRef("child1").WithParentsRef(defaultParent))
	appClient.RegisterTrainTask(client.DefaultRegisterTrainTaskOptions().WithKeyRef("child2").WithParentsRef([]string{"child1"}))

	// First task done
	appClient.StartTrainTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultTaskRef, "model0")
	appClient.DoneTrainTask(client.DefaultTaskRef)
	// second done
	appClient.StartTrainTask("child1")
	appClient.RegisterModel("child1", "model1")
	appClient.DoneTrainTask("child1")
	// last task
	appClient.StartTrainTask("child2")
	appClient.RegisterModel("child2", "model2")
	appClient.DoneTrainTask("child2")

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
