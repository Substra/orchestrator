//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
)

// TestRegisterPerformance registers a test task, start it, register its performance,
// and ensure an event containing the performance is recorded.
func TestRegisterPerformance(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.SetReadyFromWaitingFunction(client.DefaultSimpleFunctionRef)
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterFunction(client.DefaultPredictFunctionOptions())
	appClient.SetReadyFromWaitingFunction(client.DefaultPredictFunctionRef)

	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions").WithTaskOutput("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterFunction(client.DefaultMetricFunctionOptions().WithKeyRef("testmetric"))
	appClient.SetReadyFromWaitingFunction("testmetric")
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithFunctionRef("testmetric").
		WithDataSampleRef("testds").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}))
	appClient.StartTask("testTask")

	registeredPerf, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithTaskOutput("performance"))
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

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.SetReadyFromWaitingFunction(client.DefaultSimpleFunctionRef)
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterFunction(client.DefaultPredictFunctionOptions())
	appClient.SetReadyFromWaitingFunction(client.DefaultPredictFunctionRef)
	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions").WithTaskOutput("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterFunction(client.DefaultMetricFunctionOptions().WithKeyRef("testmetric"))
	appClient.SetReadyFromWaitingFunction("testmetric")
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithDataSampleRef("testds").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}).
		WithFunctionRef("testmetric"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithTaskOutput("performance"))
	require.NoError(t, err)
	appClient.DoneTask("testTask")

	task := appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)
}

func TestRegisterMultiplePerformancesForSameTaskOutput(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.SetReadyFromWaitingFunction(client.DefaultSimpleFunctionRef)
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterFunction(client.DefaultPredictFunctionOptions())
	appClient.SetReadyFromWaitingFunction(client.DefaultPredictFunctionRef)

	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions").WithTaskOutput("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterFunction(client.DefaultMetricFunctionOptions().WithKeyRef("testmetric"))
	appClient.SetReadyFromWaitingFunction("testmetric")
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithFunctionRef("testmetric").
		WithDataSampleRef("testds").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithTaskOutput("performance"))
	require.NoError(t, err)

	appClient.DoneTask("testTask")
	task := appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)

	_, err = appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithTaskOutput("performance"))
	require.ErrorContains(t, err, orcerrors.ErrBadRequest)

	task = appClient.GetComputeTask("testTask")
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)
}

func TestQueryPerformances(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.SetReadyFromWaitingFunction(client.DefaultSimpleFunctionRef)
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	appClient.RegisterFunction(client.DefaultPredictFunctionOptions().WithKeyRef("predictFunction"))
	appClient.SetReadyFromWaitingFunction("predictFunction")
	appClient.RegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("predictTask").
		WithDataSampleRef("testds").
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}).
		WithFunctionRef("predictFunction"))
	appClient.StartTask("predictTask")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("predictTask").WithKeyRef("predictions").WithTaskOutput("predictions"))
	appClient.DoneTask("predictTask")

	appClient.RegisterFunction(client.DefaultMetricFunctionOptions().WithKeyRef("testmetric"))
	appClient.SetReadyFromWaitingFunction("testmetric")
	appClient.RegisterTasks(client.DefaultTestTaskOptions().
		WithKeyRef("testTask").
		WithDataSampleRef("testds").
		WithInput("predictions", &client.TaskOutputRef{TaskRef: "predictTask", Identifier: "predictions"}).
		WithFunctionRef("testmetric"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(
		client.DefaultPerformanceOptions().WithTaskRef("testTask").WithTaskOutput("performance"),
	)
	require.NoError(t, err)

	res := appClient.QueryPerformances(nil, "", 10)
	performances := res.Performances

	require.GreaterOrEqual(t, len(performances), 1)
	require.LessOrEqual(t, performances[0].CreationDate.AsTime(), performances[1].CreationDate.AsTime())
}
