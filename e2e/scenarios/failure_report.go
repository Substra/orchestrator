package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/golang/protobuf/proto"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var failureReportScenarios = []Scenario{
	{
		testRegisterFailureReport,
		[]string{"short", "failure", "report"},
	},
}

// Register a task, start it, fail it, register a failure report on it,
// and ensure an event containing the failure report is recorded.
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

	registeredFailureReport := appClient.RegisterFailureReport(client.DefaultTaskRef)
	task := appClient.GetComputeTask(client.DefaultTaskRef)

	if registeredFailureReport.ComputeTaskKey != task.Key {
		log.WithField("task key", client.DefaultTaskRef).WithField("registeredFailureReport", registeredFailureReport).Fatal("Task keys don't match")
	}
	if task.Status != asset.ComputeTaskStatus_STATUS_FAILED {
		log.Fatal("compute task should be FAILED")
	}

	retrievedFailureReport := appClient.GetFailureReport(client.DefaultTaskRef)
	if !proto.Equal(registeredFailureReport, retrievedFailureReport) {
		log.WithField("registeredFailureReport", registeredFailureReport).WithField("retrievedFailureReport", retrievedFailureReport).
			Fatal("The retrieved failure report differs from the retrieved failure report")
	}

	eventResp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredFailureReport.ComputeTaskKey,
		AssetKind: asset.AssetKind_ASSET_FAILURE_REPORT,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	if len(eventResp.Events) != 1 {
		log.Fatalf("Unexpected number of events. Expected 1, got %d", len(eventResp.Events))
	}

	eventFailureReport := eventResp.Events[0].GetFailureReport()
	if !proto.Equal(registeredFailureReport, eventFailureReport) {
		log.WithField("registeredFailureReport", registeredFailureReport).
			WithField("eventFailureReport", eventFailureReport).
			Fatal("The failure report in the event differs from the registered failure report")
	}
}
