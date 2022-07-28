//go:build e2e
// +build e2e

package e2e

import (
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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterAlgo(client.DefaultPredictAlgoOptions())

	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithParentsRef(client.DefaultTrainTaskRef).
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions().WithKeyRef("testmetric"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithAlgoRef("testmetric").
		WithDataSampleRef("testds").
		WithParentsRef("predictTask").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}))
	appClient.StartTask("testTask")

	registeredPerf, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"))
	require.NoError(t, err)

	appClient.DoneTask("testTask")
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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterAlgo(client.DefaultPredictAlgoOptions())
	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithParentsRef(client.DefaultTrainTaskRef).
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions().WithKeyRef("testmetric"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithDataSampleRef("testds").
		WithParentsRef("predictTask").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}).
		WithAlgoRef("testmetric"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"))
	require.NoError(t, err)
	appClient.DoneTask("testTask")

	task := appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)
}

func TestRegisterMultiplePerformancesForSameMetric(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterAlgo(client.DefaultPredictAlgoOptions())

	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithParentsRef(client.DefaultTrainTaskRef).
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions().WithKeyRef("testmetric"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithAlgoRef("testmetric").
		WithDataSampleRef("testds").
		WithParentsRef("predictTask").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"))
	require.NoError(t, err)

	appClient.DoneTask("testTask")
	task := appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)

	_, err = appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"))
	require.ErrorContains(t, err, orcerrors.ErrBadRequest)

	task = appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)
}

func TestQueryPerformances(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterAlgo(client.DefaultPredictAlgoOptions().WithKeyRef("predictAlgo"))
	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithParentsRef(client.DefaultTrainTaskRef).
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}).
		WithAlgoRef("predictAlgo"))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions().WithKeyRef("testmetric"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithDataSampleRef("testds").
		WithParentsRef("predictTask").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}).
		WithAlgoRef("testmetric"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(
		client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"),
	)
	require.NoError(t, err)

	res := appClient.QueryPerformances(nil, "", 10)
	performances := res.Performances

	require.GreaterOrEqual(t, len(performances), 1)
	require.LessOrEqual(t, performances[0].CreationDate.AsTime(), performances[1].CreationDate.AsTime())
}
