//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	e2erequire "github.com/owkin/orchestrator/e2e/require"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/require"
)

// TestRegisterModel registers a task, start it, register a model on it,
// and ensure an event containing the model is recorded.
func TestRegisterModel(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	plan := appClient.GetComputePlan("cp")
	require.EqualValues(t, 1, plan.TaskCount)

	appClient.StartTask(client.DefaultTrainTaskRef)
	registeredModel := appClient.RegisterModel(client.DefaultModelOptions())

	taskEvents := appClient.QueryEvents(&asset.EventQueryFilter{AssetKey: appClient.GetKeyStore().GetKey(client.DefaultTrainTaskRef)}, "", 10)

	// 3 events: creation, start, done
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

// TestDeleteIntermediary registers 3 successive tasks, start and register models then check for model deletion
func TestDeleteIntermediary(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithDeleteIntermediaryModels(true))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child1").WithParentsRef(client.DefaultTrainTaskRef))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child2").WithParentsRef("child1"))

	// First task done
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model0"))
	// second done
	appClient.StartTask("child1")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model1").WithTaskRef("child1"))
	// last task
	appClient.StartTask("child2")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model2").WithTaskRef("child2"))

	models := appClient.GetTaskOutputModels(client.DefaultTrainTaskRef)
	require.Len(t, models, 1, "invalid number of output models")
	require.NotNil(t, models[0].Address)
	require.True(t, appClient.CanDisableModel("model0"), "parent model cannot be disabled")
	require.False(t, appClient.CanDisableModel("model2"), "final model can be disabled")

	appClient.DisableModel("model0")
	models = appClient.GetTaskOutputModels(client.DefaultTrainTaskRef)
	require.Nil(t, models[0].Address, "model has not been disabled")

	_, err := appClient.FailableRegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("badinput").WithParentsRef(client.DefaultTrainTaskRef))
	require.ErrorContains(t, err, "OE0101", "registering a task with disabled input models should fail")

	log.WithError(err).Debug("Failed to register task, as expected")
}

func TestRegisterTwoSimpleModelsForTrainTask(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.StartTask(client.DefaultTrainTaskRef)
	_, err := appClient.FailableRegisterModels(
		client.DefaultModelOptions().WithKeyRef("mod1"),
		client.DefaultModelOptions().WithKeyRef("mod2"),
	)

	require.ErrorContains(t, err, "OE0003")
	log.WithError(err).Debug("Failed to register models, as expected")
}

func TestRegisterAllModelsForCompositeTask(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultCompositeAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultCompositeTaskOptions())

	appClient.StartTask(client.DefaultCompositeTaskRef)
	appClient.RegisterModels(
		client.DefaultModelOptions().WithTaskRef(client.DefaultCompositeTaskRef).WithCategory(asset.ModelCategory_MODEL_HEAD).WithKeyRef("mod1"),
		client.DefaultModelOptions().WithTaskRef(client.DefaultCompositeTaskRef).WithCategory(asset.ModelCategory_MODEL_SIMPLE).WithKeyRef("mod2"),
	)

	task := appClient.GetComputeTask(client.DefaultCompositeTaskRef)
	require.Equal(t, asset.ComputeTaskStatus_STATUS_DONE, task.Status)
}
