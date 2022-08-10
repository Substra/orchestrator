//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/stretchr/testify/require"
)

// TestRegisterFailureReport registers a task, starts it, fails it, registers a failure report on it,
// and ensures an event containing the failure report is recorded.
func TestRegisterFailureReport(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.StartTask(client.DefaultTrainTaskRef)

	registeredFailureReport := appClient.RegisterFailureReport(client.DefaultTrainTaskRef)
	task := appClient.GetComputeTask(client.DefaultTrainTaskRef)

	require.Equal(t, task.Key, registeredFailureReport.ComputeTaskKey)
	require.Equal(t, asset.ComputeTaskStatus_STATUS_FAILED, task.Status)

	retrievedFailureReport := appClient.GetFailureReport(client.DefaultTrainTaskRef)
	e2erequire.ProtoEqual(t, registeredFailureReport, retrievedFailureReport)

	eventResp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredFailureReport.ComputeTaskKey,
		AssetKind: asset.AssetKind_ASSET_FAILURE_REPORT,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	require.Equal(t, 1, len(eventResp.Events))

	eventFailureReport := eventResp.Events[0].GetFailureReport()
	e2erequire.ProtoEqual(t, registeredFailureReport, eventFailureReport)
}
