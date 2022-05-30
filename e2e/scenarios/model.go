package scenarios

import (
	"strings"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/protobuf/proto"
)

var modelTestScenarios = []Scenario{
	{
		testRegisterModel,
		[]string{"short", "model"},
	},
	{
		testRegisterTwoSimpleModelsForTrainTask,
		[]string{"short", "model"},
	},
	{
		testRegisterAllModelsForCompositeTask,
		[]string{"short", "model", "composite"},
	},
	{
		testDeleteIntermediary,
		[]string{"short", "model"},
	},
}

// Register a task, start it, register a model on it,
// and ensure an event containing the model is recorded.
func testRegisterModel(factory *client.TestClientFactory) {
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
	registeredModel := appClient.RegisterModel(client.DefaultModelOptions())

	taskEvents := appClient.QueryEvents(&asset.EventQueryFilter{AssetKey: appClient.GetKeyStore().GetKey(client.DefaultTaskRef)}, "", 10)

	if len(taskEvents.Events) != 3 {
		// 3 events: creation, start, done
		log.WithField("events", taskEvents.Events).Fatal("Unexpected number of events")
	}

	retrievedModel := appClient.GetModel(client.DefaultModelRef)
	if !proto.Equal(registeredModel, retrievedModel) {
		log.WithField("registeredModel", registeredModel).WithField("retrievedModel", retrievedModel).
			Fatal("The retrieved model differs from the registered model")
	}

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredModel.Key,
		AssetKind: asset.AssetKind_ASSET_MODEL,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	if len(resp.Events) != 1 {
		log.Fatalf("Unexpected number of events. Expected 1, got %d", len(resp.Events))
	}

	eventModel := resp.Events[0].GetModel()
	if !proto.Equal(registeredModel, eventModel) {
		log.WithField("registeredModel", registeredModel).WithField("eventModel", eventModel).
			Fatal("The model in the event differs from the registered model")
	}
}

// register 3 successive tasks, start and register models then check for model deletion
func testDeleteIntermediary(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithDeleteIntermediaryModels(true))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child1").WithParentsRef(client.DefaultTaskRef))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child2").WithParentsRef("child1"))

	// First task done
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model0"))
	// second done
	appClient.StartTask("child1")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model1").WithTaskRef("child1"))
	// last task
	appClient.StartTask("child2")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model2").WithTaskRef("child2"))

	models := appClient.GetTaskOutputModels(client.DefaultTaskRef)
	if len(models) != 1 {
		log.Fatal("invalid number of output models")
	}

	if models[0].Address == nil {
		log.Fatal("model has a invalid address")
	}

	if !appClient.CanDisableModel("model0") {
		log.Fatal("parent model cannot be disabled")
	}

	if appClient.CanDisableModel("model2") {
		log.Fatal("final model can be disabled")
	}

	appClient.DisableModel("model0")
	models = appClient.GetTaskOutputModels(client.DefaultTaskRef)
	if models[0].Address != nil {
		log.Fatal("model has not been disabled")
	}

	_, err := appClient.FailableRegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("badinput").WithParentsRef(client.DefaultTaskRef))
	if err == nil {
		log.Fatal("registering a task with disabled input models should fail")
	}
	if !strings.Contains(err.Error(), "OE0101") {
		log.WithError(err).Fatal("Unexpected error code")
	}
	log.WithError(err).Debug("Failed to register task, as expected")
}

func testRegisterTwoSimpleModelsForTrainTask(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.StartTask(client.DefaultTaskRef)
	_, err := appClient.FailableRegisterModels(
		client.DefaultModelOptions().WithKeyRef("mod1"),
		client.DefaultModelOptions().WithKeyRef("mod2"),
	)

	if err == nil {
		log.Fatal("Model registration should have failed")
	}
	if !strings.Contains(err.Error(), "OE0003") {
		log.WithError(err).Fatal("Unexpected error code")
	}
	log.WithError(err).Debug("Failed to register models, as expected")
}

func testRegisterAllModelsForCompositeTask(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_COMPOSITE))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultCompositeTaskOptions())

	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModels(
		client.DefaultModelOptions().WithCategory(asset.ModelCategory_MODEL_HEAD).WithKeyRef("mod1"),
		client.DefaultModelOptions().WithCategory(asset.ModelCategory_MODEL_SIMPLE).WithKeyRef("mod2"),
	)

	task := appClient.GetComputeTask(client.DefaultTaskRef)
	if task.Status != asset.ComputeTaskStatus_STATUS_DONE {
		log.WithField("task", task).Fatal("Task should be DONE")
	}
}
