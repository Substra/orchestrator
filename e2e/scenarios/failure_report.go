package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var failureReportScenarios = []Scenario{
	{
		testRegisterFailureReport,
		[]string{"short", "failure", "report"},
	},
}

// Register a task, start it, fail it, and register a failure report on it.
func testRegisterFailureReport(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	plan := appClient.GetComputePlan("cp")
	if plan.TaskCount != 1 {
		log.Fatal("compute plan has invalid task count")
	}

	appClient.StartTask(client.DefaultTaskRef)

	resp := appClient.RegisterFailureReport(client.DefaultTaskRef)
	task := appClient.GetComputeTask(client.DefaultTaskRef)

	if resp.ComputeTaskKey != task.Key {
		log.WithField("task key", client.DefaultTaskRef).WithField("failureReport", resp).Fatal("Task keys don't match")
	}
	if task.Status != asset.ComputeTaskStatus_STATUS_FAILED {
		log.Fatal("compute task should be FAILED")
	}

	failureReport := appClient.GetFailureReport(client.DefaultTaskRef)
	if failureReport.ComputeTaskKey != task.Key {
		log.WithField("task key", client.DefaultTaskRef).WithField("failureReport", failureReport).Fatal("Task keys don't match")
	}
}
