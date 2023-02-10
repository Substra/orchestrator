//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/utils"
)

// TestRegisterComputePlan registers a compute plan and ensure an event containing the compute plan is recorded.
func TestRegisterComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()
	registeredPlan := appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	retrievedPlan := appClient.GetComputePlan(client.DefaultPlanRef)

	e2erequire.ProtoEqual(t, registeredPlan, retrievedPlan)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredPlan.Key,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	require.Equal(t, len(resp.Events), 1, "Unexpected number of events")

	eventPlan := resp.Events[0].GetComputePlan()
	e2erequire.ProtoEqual(t, registeredPlan, eventPlan)
}

func TestCancelComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultCompositeFunctionOptions().WithKeyRef("compFunction"))
	appClient.RegisterFunction(client.DefaultAggregateFunctionOptions().WithKeyRef("aggFunction"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().WithKeyRef("cmp1").WithFunctionRef("compFunction"))
	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().WithKeyRef("cmp2").WithFunctionRef("compFunction"))

	appClient.RegisterTasks(client.DefaultAggregateTaskOptions().
		WithKeyRef("agg1").
		WithFunctionRef("aggFunction").
		WithInput("model", &client.TaskOutputRef{TaskRef: "cmp1", Identifier: "shared"}).
		WithInput("model", &client.TaskOutputRef{TaskRef: "cmp2", Identifier: "shared"}))

	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().
		WithKeyRef("cmp3").
		WithFunctionRef("compFunction").
		WithInput("local", &client.TaskOutputRef{TaskRef: "cmp1", Identifier: "local"}).
		WithInput("shared", &client.TaskOutputRef{TaskRef: "agg1", Identifier: "model"}))

	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().
		WithKeyRef("cmp4").
		WithFunctionRef("compFunction").
		WithInput("local", &client.TaskOutputRef{TaskRef: "cmp2", Identifier: "local"}).
		WithInput("shared", &client.TaskOutputRef{TaskRef: "agg1", Identifier: "model"}))

	// We start processing the compute plan
	appClient.StartTask("cmp1")
	appClient.StartTask("cmp2")

	appClient.RegisterModels(
		client.DefaultModelOptions().WithTaskRef("cmp1").WithTaskOutput("local").WithKeyRef("cmp1h"),
		client.DefaultModelOptions().WithTaskRef("cmp1").WithTaskOutput("shared").WithKeyRef("cmp1s"),
	)

	// initially, the cp is not canceled
	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Nil(t, plan.FailureDate)
	require.Nil(t, plan.CancelationDate)

	// we cancel the cp
	_, err := appClient.CancelComputePlan(client.DefaultPlanRef)
	require.NoError(t, err)

	plan = appClient.GetComputePlan(client.DefaultPlanRef)
	require.Nil(t, plan.FailureDate)
	require.NotNil(t, plan.CancelationDate)

	// we cannot cancel the cp a second time
	_, err = appClient.CancelComputePlan(client.DefaultPlanRef)
	require.Errorf(t, err, "already terminated")
}

// TestMultiStageComputePlan is the "canonical" example of FL with 2 organizations aggregating their trunks
// This does not check multi-organization setup though!
//
//	 ,========,                ,========,
//	 | ORG A  |                | ORG B  |
//	 *========*                *========*
//
//	   ø     ø                  ø      ø
//	   |     |                  |      |
//	   hd    tr                 tr     hd
//	 -----------              -----------
//	| Composite |            | Composite |      STEP 1
//	 -----------              -----------
//	   hd    tr                 tr     hd
//	   |      \   ,========,   /      |
//	   |       \  | ORG C  |  /       |
//	   |        \ *========* /        |
//	   |       ----------------       |
//	   |      |    Aggregate   |      |         STEP 2
//	   |       ----------------       |
//	   |              |               |
//	   |     ,_______/ \_______       |
//	   |     |                 |      |
//	  hd    tr                tr     hd
//	 -----------             -----------
//	| Composite |           | Composite |       STEP 3
//	 -----------             -----------
//	  hd    tr                 tr     hd
//	          \                /
//	           \              /
//	            \            /
//	           ----------------
//	          |    Aggregate   |                STEP 4
//	           ----------------
func TestMultiStageComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultCompositeFunctionOptions().WithKeyRef("functionComp"))
	appClient.RegisterFunction(client.DefaultAggregateFunctionOptions().WithKeyRef("functionAgg"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	// step 1
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compA1").WithFunctionRef("functionComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compB1").WithFunctionRef("functionComp"),
	)
	// step 2
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().
			WithKeyRef("aggC2").
			WithInput("model", &client.TaskOutputRef{TaskRef: "compA1", Identifier: "shared"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "compB1", Identifier: "shared"}).
			WithFunctionRef("functionAgg"),
	)
	// step 3
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("compA3").
			WithInput("local", &client.TaskOutputRef{TaskRef: "compA1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "aggC2", Identifier: "model"}).
			WithFunctionRef("functionComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("compB3").
			WithInput("local", &client.TaskOutputRef{TaskRef: "compB1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "aggC2", Identifier: "model"}).
			WithFunctionRef("functionComp"),
	)
	// step 4
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().
			WithKeyRef("aggC4").
			WithInput("model", &client.TaskOutputRef{TaskRef: "compA3", Identifier: "shared"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "compB3", Identifier: "shared"}).
			WithFunctionRef("functionAgg"),
	)

	lastAggregate := appClient.GetComputeTask("aggC4")
	if lastAggregate.Rank != 3 {
		log.Fatal().Int32("rank", lastAggregate.Rank).Msg("last aggegation task has not expected rank")
	}

	// Start step 1
	appClient.StartTask("compA1")
	appClient.StartTask("compB1")

	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA1").
			WithKeyRef("modelA1H").
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA1").
			WithKeyRef("modelA1T").
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compA1")
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB1").
			WithKeyRef("modelB1H").
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB1").
			WithKeyRef("modelB1T").
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compB1")

	// Start step 2
	appClient.StartTask("aggC2")
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("aggC2").
			WithKeyRef("modelC2"),
	)
	appClient.DoneTask("aggC2")

	// Start step 3
	appClient.StartTask("compA3")
	appClient.StartTask("compB3")

	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA3").
			WithKeyRef("modelA3H").
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA3").
			WithKeyRef("modelA3T").
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compA3")
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB3").
			WithKeyRef("modelB3H").
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB3").
			WithKeyRef("modelB3T").
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compB3")

	// Start step 4
	appClient.StartTask("aggC4")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("aggC4").WithKeyRef("modelC4"))
	appClient.DoneTask("aggC4")
}

func TestLargeComputePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	appClient := factory.NewTestClient()

	nbTasks := 10000
	nbQuery := 5000 // 10k exceed max response size

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	start := time.Now()
	for i := 0; i < nbTasks; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().
			WithKeyRef(fmt.Sprintf("task%d", i)).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}
	log.Info().Dur("registrationDuration", time.Since(start)).Int("nbTasks", nbTasks).Msg("registration done")

	start = time.Now()
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{FunctionKey: appClient.GetKeyStore().GetKey(client.DefaultSimpleFunctionRef)}, "", nbQuery)
	log.Info().Dur("queryDuration", time.Since(start)).Int("nbTasks", nbQuery).Msg("query done")

	require.Equal(t, nbQuery, len(resp.Tasks), "unexpected task count")
}

func TestBatchLargeComputePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	appClient := factory.NewTestClient()

	nbTasks := 10000
	batchSize := 1000
	nbQuery := 5000 // 10k exceed max response size

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	start := time.Now()
	for i := 0; i < nbTasks; {
		batchStart := time.Now()
		newTasks := make([]client.Taskable, 0, batchSize)
		for c := 0; c < batchSize && i < nbTasks; c++ {
			i++
			newTasks = append(newTasks, client.DefaultTrainTaskOptions().
				WithKeyRef(fmt.Sprintf("task%d", i)).
				WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
		}
		appClient.RegisterTasks(newTasks...)
		log.Info().Dur("batchDuration", time.Since(batchStart)).Int("nbTasks", i).Msg("batch done")
	}
	log.Info().Dur("registrationDuration", time.Since(start)).Int("nbTasks", nbTasks).Msg("registration done")

	start = time.Now()
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{FunctionKey: appClient.GetKeyStore().GetKey(client.DefaultSimpleFunctionRef)}, "", nbQuery)
	log.Info().Dur("queryDuration", time.Since(start)).Int("nbTasks", nbQuery).Msg("query done")

	require.Equal(t, nbQuery, len(resp.Tasks), "unexpected task count")
}

func TestSmallComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterFunction(client.DefaultPredictFunctionOptions())
	appClient.RegisterFunction(client.DefaultMetricFunctionOptions())

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),

		client.DefaultTrainTaskOptions().WithKeyRef("train3").
			WithInput("model", &client.TaskOutputRef{TaskRef: "train1", Identifier: "model"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "train2", Identifier: "model"}),

		client.DefaultPredictTaskOptions().WithKeyRef("predict").
			WithInput("model", &client.TaskOutputRef{TaskRef: "train3", Identifier: "model"}),

		client.DefaultTestTaskOptions().
			WithDataSampleRef("objSample").
			WithInput("predictions", &client.TaskOutputRef{TaskRef: "predict", Identifier: "predictions"}),
	)
}

func TestAggregateComposite(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultCompositeFunctionOptions())
	appClient.RegisterFunction(client.DefaultAggregateFunctionOptions().WithKeyRef("aggFunction"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterFunction(client.DefaultMetricFunctionOptions())

	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("c1"),
		client.DefaultCompositeTaskOptions().WithKeyRef("c2"),

		client.
			DefaultAggregateTaskOptions().
			WithKeyRef("a1").
			WithFunctionRef("aggFunction").
			WithInput("model", &client.TaskOutputRef{TaskRef: "c1", Identifier: "shared"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "c2", Identifier: "shared"}),

		client.
			DefaultCompositeTaskOptions().
			WithKeyRef("c3").
			WithInput("shared", &client.TaskOutputRef{TaskRef: "a1", Identifier: "model"}).
			WithInput("local", &client.TaskOutputRef{TaskRef: "c1", Identifier: "local"}))

	appClient.StartTask("c1")
	models := []*client.ModelOptions{
		client.DefaultModelOptions().WithTaskRef("c1").WithKeyRef("m1H").WithTaskOutput("local"),
		client.DefaultModelOptions().WithTaskRef("c1").WithKeyRef("m1T").WithTaskOutput("shared"),
	}
	appClient.RegisterModels(models...)
	appClient.DoneTask("c1")

	appClient.StartTask("c2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c2").WithKeyRef("m2H").WithTaskOutput("local"))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c2").WithKeyRef("m2T").WithTaskOutput("shared"))
	appClient.DoneTask("c2")

	appClient.StartTask("a1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("a1").WithKeyRef("mAgg"))
	appClient.DoneTask("a1")

	appClient.StartTask("c3")

	inputs := appClient.GetTaskInputAssets("c3")
	inputModels := utils.Filter(inputs, func(input *asset.ComputeTaskInputAsset) bool {
		return input.GetModel() != nil
	})
	assert.Len(t, inputModels, 2)
}

func TestFailLargeComputePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	appClient := factory.NewTestClient()

	nbRounds := 1000
	nbPharma := 11
	var nbTasks int

	appClient.RegisterFunction(client.DefaultCompositeFunctionOptions().WithKeyRef("functionComp"))
	appClient.RegisterFunction(client.DefaultAggregateFunctionOptions().WithKeyRef("functionAgg"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	newTasks := make([]client.Taskable, 0)
	start := time.Now()
	for i := 0; i < nbRounds; {
		compKeys := make([]string, nbPharma)

		for pharma := 1; pharma < nbPharma+1; {
			compKey := fmt.Sprintf("compP%dR%d", pharma, i)
			compKeys[pharma-1] = compKey

			task := client.DefaultCompositeTaskOptions().WithKeyRef(compKey).WithFunctionRef("functionComp")
			if i > 0 {
				// Reference previous composite and aggregate
				local := fmt.Sprintf("compP%dR%d", pharma, i-1)
				shared := fmt.Sprintf("aggR%d", i-1)
				task.
					WithInput("local", &client.TaskOutputRef{TaskRef: local, Identifier: "local"}).
					WithInput("shared", &client.TaskOutputRef{TaskRef: shared, Identifier: "model"})
			}
			newTasks = append(newTasks, task)
			nbTasks++
			pharma++
		}

		// Add aggregate
		agg := client.
			DefaultAggregateTaskOptions().
			WithKeyRef(fmt.Sprintf("aggR%d", i)).
			WithFunctionRef("functionAgg")
		for _, compKey := range compKeys {
			agg.WithInput("model", &client.TaskOutputRef{TaskRef: compKey, Identifier: "shared"})
		}

		newTasks = append(newTasks, agg)
		nbTasks++

		i++

		if i%20 == 0 {
			appClient.RegisterTasks(newTasks...)
			log.Debug().Int("round", i).Int("nbTasks", nbTasks).Dur("duration", time.Since(start)).Msg("Round registered")
			newTasks = make([]client.Taskable, 0)
			start = time.Now()
		}
	}

	if len(newTasks) > 0 {
		appClient.RegisterTasks(newTasks...)
		log.Debug().Int("nbTasks", nbTasks).Dur("duration", time.Since(start)).Msg("Round registered")
	}

	// Fail the composite of rank 0 on pharma1
	start = time.Now()
	appClient.StartTask("compP1R0")
	appClient.FailTask("compP1R0")
	log.Info().Dur("duration", time.Since(start)).Int("nbTasks", nbTasks).Msg("canceled compute plan")

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.NotNil(t, plan.FailureDate)
	require.Nil(t, plan.CancelationDate)
}

func TestQueryComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef("cp1"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef("cp2"))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef("cp3"))

	resp := appClient.QueryPlans(&asset.PlanQueryFilter{}, "", 3)
	require.Equal(t, 3, len(resp.Plans), "unexpected number of compute plans")
}

func TestGetComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	// A CP with 1 parent task and 2 child tasks
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().
		WithKeyRef("task#1").
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Nil(t, plan.FailureDate)
	require.Nil(t, plan.CancelationDate)

	appClient.StartTask(client.DefaultTrainTaskRef)

	appClient.FailTask("task#1")
	plan = appClient.GetComputePlan(client.DefaultPlanRef)
	require.NotNil(t, plan.FailureDate)
	require.Nil(t, plan.CancelationDate)
}

func TestCompositeParentChild(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultCompositeFunctionOptions().WithKeyRef("functionComp"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("comp1").WithFunctionRef("functionComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("comp2").
			WithFunctionRef("functionComp").
			WithInput("local", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "shared"}))

	appClient.StartTask("comp1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1H").WithTaskOutput("local"))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1T").WithTaskOutput("shared"))
	appClient.DoneTask("comp1")

	appClient.StartTask("comp2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2H").WithTaskOutput("local"))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2T").WithTaskOutput("shared"))
	appClient.DoneTask("comp2")

	// Register a composite task with 2 composite parents
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("comp3").
			WithFunctionRef("functionComp").
			WithInput("local", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "comp2", Identifier: "shared"}))

	inputs := appClient.GetTaskInputAssets("comp3")
	models := utils.Filter(inputs, func(input *asset.ComputeTaskInputAsset) bool {
		return input.GetModel() != nil
	})
	require.Len(t, models, 2, "composite task should have 2 input models")
	require.Equal(t, appClient.GetKeyStore().GetKey("model1H"), models[0].GetModel().Key, "first model should be HEAD from comp1")
	require.Equal(t, appClient.GetKeyStore().GetKey("model2T"), models[1].GetModel().Key, "second model should be TRUNK from comp2")
}

// TestUpdateComputePlan updates mutable fieds of a compute plan and ensure an event containing the compute plan is recorded. List of mutable fields: name.
func TestUpdateComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()
	keyRef := "compute_plan_update"
	registeredPlan := appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef(keyRef))
	expectedPlan := appClient.GetComputePlan(keyRef)

	appClient.UpdateComputePlan(keyRef, "new compute plan name")
	retrievedPlan := appClient.GetComputePlan(keyRef)

	expectedPlan.Name = "new compute plan name"

	e2erequire.ProtoEqual(t, expectedPlan, retrievedPlan)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredPlan.Key,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
	}, "", 100)

	require.Len(t, resp.Events, 1, "Unexpected number of events")

	eventComputePlan := resp.Events[0].GetComputePlan()

	e2erequire.ProtoEqual(t, expectedPlan, eventComputePlan)
}

func TestDisableTransientOutput(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithOutput("model", &asset.NewPermissions{Public: true}, true))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("child1").WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))

	// First task done
	appClient.StartTask(client.DefaultTrainTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model0"))
	appClient.DoneTask(client.DefaultTrainTaskRef)
	// second done
	appClient.StartTask("child1")
	appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model1").WithTaskRef("child1"))
	appClient.DoneTask("child1")

	appClient.DisableOutput(client.DefaultTrainTaskRef, "model")
	models := appClient.GetTaskOutputModels(client.DefaultTrainTaskRef)
	require.Nil(t, models[0].Address, "model has not been disabled")

	_, err := appClient.FailableRegisterTasks(client.DefaultPredictTaskOptions().
		WithKeyRef("badinput").
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))

	require.ErrorContains(t, err, "OE0101", "registering a task with disabled input models should fail")

}

// TestIsPlanRunning ensures that the compute plan is considered as
// running when there are tasks being executed or waiting to be executed.
func TestIsPlanRunning(t *testing.T) {
	appClient := factory.NewTestClient()
	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	resp := appClient.IsPlanRunning(client.DefaultPlanRef)
	require.False(t, resp.IsRunning)

	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	resp = appClient.IsPlanRunning(client.DefaultPlanRef)
	require.True(t, resp.IsRunning)

	appClient.StartTask(client.DefaultTrainTaskRef)

	resp = appClient.IsPlanRunning(client.DefaultPlanRef)
	require.True(t, resp.IsRunning)

	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.DoneTask(client.DefaultTrainTaskRef)

	resp = appClient.IsPlanRunning(client.DefaultPlanRef)
	require.False(t, resp.IsRunning)

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithKeyRef("task2"))

	resp = appClient.IsPlanRunning(client.DefaultPlanRef)
	require.True(t, resp.IsRunning)

	appClient.FailTask("task2")

	resp = appClient.IsPlanRunning(client.DefaultPlanRef)
	require.False(t, resp.IsRunning)

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions().WithKeyRef("cp2"))
	appClient.RegisterTasks(client.DefaultTrainTaskOptions().WithPlanRef("cp2").WithKeyRef("task-cp2"))
	appClient.StartTask("task-cp2")

	_, err := appClient.CancelComputePlan("cp2")
	require.Nil(t, err)

	resp = appClient.IsPlanRunning("cp2")
	require.True(t, resp.IsRunning)

	appClient.CancelTask("task-cp2")

	resp = appClient.IsPlanRunning("cp2")
	require.False(t, resp.IsRunning)
}
