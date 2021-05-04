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
	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"google.golang.org/grpc"
)

// Register a task and its dependencies, then start the task.
func testTrainTaskLifecycle(conn *grpc.ClientConn) {
	appClient, err := client.NewTestClient(conn, "MyOrg1MSP", "test-train-task-lifecycle")
	if err != nil {
		log.WithError(err).Fatal("could not create TestClient")
	}

	appClient.RegisterAlgo()
	appClient.RegisterDataManager()
	appClient.RegisterDataSample()
	appClient.RegisterComputePlan()
	appClient.RegisterTrainTask()
	appClient.RegisterChildTask()
	appClient.StartTrainTask()
}

// register a task, start it, and register a model on it.
func testRegisterModel(conn *grpc.ClientConn) {
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
