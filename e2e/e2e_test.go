// +build e2e

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

// Package e2e contains end to end tests of the orchestrator.
// Note that due to the complexity of the distributed mode, tests only targets standalone orchestration.
package e2e

import (
	"flag"
	"os"
	"testing"
)

// orchestration app, see harness.go for more details
var app *testApp

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		// Skip e2e testing in short mode
		os.Exit(0)
	}

	app = newTestApp()
	app.listen()

	ret := m.Run()

	app.runnable.Stop()
	os.Exit(ret)
}

// This is a test scenario: it will run all steps in order
func TestTrainTaskLifecycle(t *testing.T) {
	appClient, err := newAppClient("MyOrg1Test", "test-train-task-lifecycle")
	if err != nil {
		t.Error(err)
	}
	defer appClient.Close()

	t.Run("register node", appClient.registerNode)
	t.Run("register algo", appClient.registerAlgo)
	t.Run("register datamanager", appClient.registerDataManager)
	t.Run("register datasample", appClient.registerDataSample)
	t.Run("register train task", appClient.registerTrainTask)
	t.Run("register child task", appClient.registerChildTask)
	t.Run("start train task", appClient.startTrainTask)
}

func TestRegisterModel(t *testing.T) {
	appClient, err := newAppClient("MyOrg1Test", "test-register-model")
	if err != nil {
		t.Error(err)
	}
	defer appClient.Close()

	t.Run("register node", appClient.registerNode)
	t.Run("register algo", appClient.registerAlgo)
	t.Run("register datamanager", appClient.registerDataManager)
	t.Run("register datasample", appClient.registerDataSample)
	t.Run("register train task", appClient.registerTrainTask)
	t.Run("start train task", appClient.startTrainTask)
	t.Run("register model", appClient.registerModel)
}
