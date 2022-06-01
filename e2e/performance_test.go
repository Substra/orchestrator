//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"testing"

	"github.com/owkin/orchestrator/e2e/client"
	e2erequire "github.com/owkin/orchestrator/e2e/require"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/stretchr/testify/require"
)

// TestRegisterPerformance registers a test task, start it, register its performance,
// and ensure an event containing the performance is recorded.
func TestRegisterPerformance(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithDataSampleRef("testds").WithParentsRef(client.DefaultTaskRef).WithMetricsRef("testmetric"))
	appClient.StartTask("testTask")

	registeredPerf, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"))
	require.NoError(t, err)

	task := appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)

	perfResp := appClient.QueryPerformances(&asset.PerformanceQueryFilter{
		ComputeTaskKey: task.Key,
	}, "", 100)
	require.Equal(t, 1, len(perfResp.Performances))

	retrievedPerf := perfResp.Performances[0]
	e2erequire.ProtoEqual(t, registeredPerf, retrievedPerf)

	eventResp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredPerf.GetKey(),
		AssetKind: asset.AssetKind_ASSET_PERFORMANCE,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)
	require.Equal(t, 1, len(eventResp.Events))

	eventPerf := eventResp.Events[0].GetPerformance()
	e2erequire.ProtoEqual(t, registeredPerf, eventPerf)
}

func TestRegisterMultiplePerformances(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric1"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric2"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithDataSampleRef("testds").WithParentsRef(client.DefaultTaskRef).WithMetricsRef("testmetric1", "testmetric2"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric1"))
	require.NoError(t, err)

	task := appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DOING, task.Status)

	_, err = appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric2"))
	require.NoError(t, err)

	task = appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)
}

func TestRegisterMultiplePerformancesForSameMetric(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric1"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric2"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithDataSampleRef("testds").WithParentsRef(client.DefaultTaskRef).WithMetricsRef("testmetric1", "testmetric2"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric1"))
	require.NoError(t, err)

	task := appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DOING, task.Status)

	_, err = appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric1"))
	require.ErrorContains(t, err, orcerrors.ErrConflict)

	task = appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DOING, task.Status)
}

func TestQueryPerformances(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric"))

	const nbPerformances = 2
	for i := 0; i < nbPerformances; i++ {
		testTaskRef := fmt.Sprint("testTask", i)
		appClient.RegisterTasks(client.DefaultTestTaskOptions().
			WithKeyRef(testTaskRef).
			WithDataSampleRef("testds").
			WithParentsRef(client.DefaultTaskRef).
			WithMetricsRef("testmetric"))
		appClient.StartTask(testTaskRef)

		_, err := appClient.RegisterPerformance(
			client.DefaultPerformanceOptions().WithTaskRef(testTaskRef).WithMetricRef("testmetric"),
		)
		require.NoError(t, err)
	}

	res := appClient.QueryPerformances(nil, "", 10)
	performances := res.Performances

	require.GreaterOrEqual(t, len(performances), nbPerformances)
	require.LessOrEqual(t, performances[0].CreationDate.AsTime(), performances[1].CreationDate.AsTime())
}
