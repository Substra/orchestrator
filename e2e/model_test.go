//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
)

// TestRegisterModel registers a task, start it, register a model on it,
// and ensure an event containing the model is recorded.
func TestRegisterModel(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.StartTask(client.DefaultTrainTaskRef)
	registeredModel := appClient.RegisterModel(client.DefaultModelOptions())

	taskEvents := appClient.QueryEvents(&asset.EventQueryFilter{AssetKey: appClient.GetKeyStore().GetKey(client.DefaultTrainTaskRef)}, "", 10)

	// 3 events: creation, start, task output asset creation
	require.Equalf(t, 3, len(taskEvents.Events), "events: %v", taskEvents.Events)

	retrievedModel := appClient.GetModel(client.DefaultModelRef)
	e2erequire.ProtoEqual(t, registeredModel, retrievedModel)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredModel.Key,
		AssetKind: asset.AssetKind_ASSET_MODEL,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	require.Len(t, resp.Events, 1)

	eventModel := resp.Events[0].GetModel()
	e2erequire.ProtoEqual(t, registeredModel, eventModel)
}

func TestRegisterTwoSimpleModelsForTrainTask(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.StartTask(client.DefaultTrainTaskRef)
	_, err := appClient.FailableRegisterModels(
		client.DefaultModelOptions().WithKeyRef("mod1"),
		client.DefaultModelOptions().WithKeyRef("mod2"),
	)

	require.ErrorContains(t, err, "OE0006")
	log.Debug().Err(err).Msg("Failed to register models, as expected")
}

func TestRegisterAllModelsForCompositeTask(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultCompositeFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultCompositeTaskOptions())

	appClient.StartTask(client.DefaultCompositeTaskRef)
	appClient.RegisterModels(
		client.DefaultModelOptions().WithTaskRef(client.DefaultCompositeTaskRef).WithKeyRef("mod1").WithTaskOutput("local"),
		client.DefaultModelOptions().WithTaskRef(client.DefaultCompositeTaskRef).WithKeyRef("mod2").WithTaskOutput("shared"),
	)
	appClient.DoneTask(client.DefaultCompositeTaskRef)

	task := appClient.GetComputeTask(client.DefaultCompositeTaskRef)
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)
}
