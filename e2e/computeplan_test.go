//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/stretchr/testify/require"
)

// TestRegisterComputePlan registers a compute plan and ensure an event containing the compute plan is recorded.
func TestRegisterComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()
	registeredPlan := appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	retrievedPlan := appClient.GetComputePlan(client.DefaultPlanRef)

	// Ignore dynamic fields
	retrievedPlan.Status = asset.ComputePlanStatus_PLAN_STATUS_EMPTY

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

	appClient.RegisterAlgo(client.DefaultCompositeAlgoOptions().WithKeyRef("compAlgo"))
	appClient.RegisterAlgo(client.DefaultAggregateAlgoOptions().WithKeyRef("aggAlgo"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().WithKeyRef("cmp1").WithAlgoRef("compAlgo"))
	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().WithKeyRef("cmp2").WithAlgoRef("compAlgo"))

	appClient.RegisterTasks(client.DefaultAggregateTaskOptions().
		WithKeyRef("agg1").
		WithAlgoRef("aggAlgo").
		WithParentsRef("cmp1", "cmp2").
		WithInput("model", &client.TaskOutputRef{TaskRef: "cmp1", Identifier: "shared"}).
		WithInput("model", &client.TaskOutputRef{TaskRef: "cmp2", Identifier: "shared"}))

	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().
		WithKeyRef("cmp3").
		WithAlgoRef("compAlgo").
		WithParentsRef("cmp1", "agg1").
		WithInput("local", &client.TaskOutputRef{TaskRef: "cmp1", Identifier: "local"}).
		WithInput("shared", &client.TaskOutputRef{TaskRef: "agg1", Identifier: "model"}))

	appClient.RegisterTasks(client.DefaultCompositeTaskOptions().
		WithKeyRef("cmp4").
		WithAlgoRef("compAlgo").
		WithParentsRef("cmp2", "agg1").
		WithInput("local", &client.TaskOutputRef{TaskRef: "cmp2", Identifier: "local"}).
		WithInput("shared", &client.TaskOutputRef{TaskRef: "agg1", Identifier: "model"}))

	// We start processing the compute plan
	appClient.StartTask("cmp1")
	appClient.StartTask("cmp2")

	appClient.RegisterModels(
		client.DefaultModelOptions().WithTaskRef("cmp1").WithTaskOutput("local").WithKeyRef("cmp1h").WithCategory(asset.ModelCategory_MODEL_HEAD),
		client.DefaultModelOptions().WithTaskRef("cmp1").WithTaskOutput("shared").WithKeyRef("cmp1s").WithCategory(asset.ModelCategory_MODEL_SIMPLE),
	)

	// initially, the cp is not canceled
	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.NotEqual(t, asset.ComputePlanStatus_PLAN_STATUS_CANCELED, plan.Status)
	require.Nil(t, plan.CancelationDate)

	// we cancel the cp
	appClient.CancelComputePlan(client.DefaultPlanRef)
	plan = appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_CANCELED, plan.Status)
	require.NotNil(t, plan.CancelationDate)

	// we cannot cancel the cp a second time
	_, err := appClient.CancelComputePlan(client.DefaultPlanRef)
	require.Errorf(t, err, "already canceled")
}

// TestMultiStageComputePlan is the "canonical" example of FL with 2 organizations aggregating their trunks
// This does not check multi-organization setup though!
//
//   ,========,                ,========,
//   | ORG A  |                | ORG B  |
//   *========*                *========*
//
//     ø     ø                  ø      ø
//     |     |                  |      |
//     hd    tr                 tr     hd
//   -----------              -----------
//  | Composite |            | Composite |      STEP 1
//   -----------              -----------
//     hd    tr                 tr     hd
//     |      \   ,========,   /      |
//     |       \  | ORG C  |  /       |
//     |        \ *========* /        |
//     |       ----------------       |
//     |      |    Aggregate   |      |         STEP 2
//     |       ----------------       |
//     |              |               |
//     |     ,_______/ \_______       |
//     |     |                 |      |
//     hd    tr                tr     hd
//   -----------             -----------
//  | Composite |           | Composite |       STEP 3
//   -----------             -----------
//     hd    tr                 tr     hd
//            \                /
//             \              /
//              \            /
//             ----------------
//            |    Aggregate   |                STEP 4
//             ----------------
//
//
func TestMultiStageComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultCompositeAlgoOptions().WithKeyRef("algoComp"))
	appClient.RegisterAlgo(client.DefaultAggregateAlgoOptions().WithKeyRef("algoAgg"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	// step 1
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compA1").WithAlgoRef("algoComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("compB1").WithAlgoRef("algoComp"),
	)
	// step 2
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().
			WithKeyRef("aggC2").
			WithParentsRef("compA1", "compB1").
			WithInput("model", &client.TaskOutputRef{TaskRef: "compA1", Identifier: "shared"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "compB1", Identifier: "shared"}).
			WithAlgoRef("algoAgg"),
	)
	// step 3
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("compA3").
			WithParentsRef("compA1", "aggC2").
			WithInput("local", &client.TaskOutputRef{TaskRef: "compA1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "aggC2", Identifier: "model"}).
			WithAlgoRef("algoComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("compB3").
			WithParentsRef("compB1", "aggC2").
			WithInput("local", &client.TaskOutputRef{TaskRef: "compB1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "aggC2", Identifier: "model"}).
			WithAlgoRef("algoComp"),
	)
	// step 4
	appClient.RegisterTasks(
		client.DefaultAggregateTaskOptions().
			WithKeyRef("aggC4").
			WithParentsRef("compA3", "compB3").
			WithInput("model", &client.TaskOutputRef{TaskRef: "compA3", Identifier: "shared"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "compB3", Identifier: "shared"}).
			WithAlgoRef("algoAgg"),
	)

	lastAggregate := appClient.GetComputeTask("aggC4")
	if lastAggregate.Rank != 3 {
		log.WithField("rank", lastAggregate.Rank).Fatal("last aggegation task has not expected rank")
	}

	// Start step 1
	appClient.StartTask("compA1")
	appClient.StartTask("compB1")

	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA1").
			WithKeyRef("modelA1H").
			WithCategory(asset.ModelCategory_MODEL_HEAD).
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA1").
			WithKeyRef("modelA1T").
			WithCategory(asset.ModelCategory_MODEL_SIMPLE).
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compA1")
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB1").
			WithKeyRef("modelB1H").
			WithCategory(asset.ModelCategory_MODEL_HEAD).
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB1").
			WithKeyRef("modelB1T").
			WithCategory(asset.ModelCategory_MODEL_SIMPLE).
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compB1")
	cp := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_DOING, cp.Status)

	// Start step 2
	appClient.StartTask("aggC2")
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("aggC2").
			WithKeyRef("modelC2").
			WithCategory(asset.ModelCategory_MODEL_SIMPLE),
	)
	appClient.DoneTask("aggC2")

	// Start step 3
	appClient.StartTask("compA3")
	appClient.StartTask("compB3")

	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA3").
			WithKeyRef("modelA3H").
			WithCategory(asset.ModelCategory_MODEL_HEAD).
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compA3").
			WithKeyRef("modelA3T").
			WithCategory(asset.ModelCategory_MODEL_SIMPLE).
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compA3")
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB3").
			WithKeyRef("modelB3H").
			WithCategory(asset.ModelCategory_MODEL_HEAD).
			WithTaskOutput("local"),
	)
	appClient.RegisterModel(
		client.DefaultModelOptions().
			WithTaskRef("compB3").
			WithKeyRef("modelB3T").
			WithCategory(asset.ModelCategory_MODEL_SIMPLE).
			WithTaskOutput("shared"),
	)
	appClient.DoneTask("compB3")

	// Start step 4
	appClient.StartTask("aggC4")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("aggC4").WithKeyRef("modelC4").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
	appClient.DoneTask("aggC4")
}

func TestLargeComputePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	appClient := factory.NewTestClient()

	nbTasks := 10000
	nbQuery := 5000 // 10k exceed max response size

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	start := time.Now()
	for i := 0; i < nbTasks; i++ {
		appClient.RegisterTasks(client.DefaultTrainTaskOptions().
			WithKeyRef(fmt.Sprintf("task%d", i)).
			WithParentsRef(client.DefaultTrainTaskRef).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
	}
	log.WithField("registrationDuration", time.Since(start)).WithField("nbTasks", nbTasks).Info("registration done")

	start = time.Now()
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultSimpleAlgoRef)}, "", nbQuery)
	log.WithField("queryDuration", time.Since(start)).WithField("nbTasks", nbQuery).Info("query done")

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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
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
				WithParentsRef(client.DefaultTrainTaskRef).
				WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
		}
		appClient.RegisterTasks(newTasks...)
		log.WithField("batchDuration", time.Since(batchStart)).WithField("nbTasks", i).Info("batch done")
	}
	log.WithField("registrationDuration", time.Since(start)).WithField("nbTasks", nbTasks).Info("registration done")

	start = time.Now()
	resp := appClient.QueryTasks(&asset.TaskQueryFilter{AlgoKey: appClient.GetKeyStore().GetKey(client.DefaultSimpleAlgoRef)}, "", nbQuery)
	log.WithField("queryDuration", time.Since(start)).WithField("nbTasks", nbQuery).Info("query done")

	require.Equal(t, nbQuery, len(resp.Tasks), "unexpected task count")
}

func TestSmallComputePlan(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterAlgo(client.DefaultPredictAlgoOptions())
	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions())

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),

		client.DefaultTrainTaskOptions().WithKeyRef("train3").
			WithParentsRef("train1", "train2").
			WithInput("model", &client.TaskOutputRef{TaskRef: "train1", Identifier: "model"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "train2", Identifier: "model"}),

		client.DefaultPredictTaskOptions().WithKeyRef("predict").
			WithParentsRef("train3").
			WithInput("model", &client.TaskOutputRef{TaskRef: "train3", Identifier: "model"}),

		client.DefaultTestTaskOptions().
			WithDataSampleRef("objSample").
			WithParentsRef("predict").
			WithInput("predictions", &client.TaskOutputRef{TaskRef: "predict", Identifier: "predictions"}),
	)

	cp := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_TODO, cp.Status, "unexpected plan status")
}

func TestAggregateComposite(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultCompositeAlgoOptions())
	appClient.RegisterAlgo(client.DefaultAggregateAlgoOptions().WithKeyRef("aggAlgo"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions())

	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("c1"),
		client.DefaultCompositeTaskOptions().WithKeyRef("c2"),

		client.
			DefaultAggregateTaskOptions().
			WithKeyRef("a1").
			WithAlgoRef("aggAlgo").
			WithParentsRef("c1", "c2").
			WithInput("model", &client.TaskOutputRef{TaskRef: "c1", Identifier: "shared"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "c2", Identifier: "shared"}),

		client.
			DefaultCompositeTaskOptions().
			WithKeyRef("c3").
			WithParentsRef("a1", "c1").
			WithInput("shared", &client.TaskOutputRef{TaskRef: "a1", Identifier: "model"}).
			WithInput("local", &client.TaskOutputRef{TaskRef: "c1", Identifier: "local"}))

	appClient.StartTask("c1")
	models := []*client.ModelOptions{
		client.DefaultModelOptions().WithTaskRef("c1").WithKeyRef("m1H").WithCategory(asset.ModelCategory_MODEL_HEAD).WithTaskOutput("local"),
		client.DefaultModelOptions().WithTaskRef("c1").WithKeyRef("m1T").WithCategory(asset.ModelCategory_MODEL_SIMPLE).WithTaskOutput("shared"),
	}
	appClient.RegisterModels(models...)
	appClient.DoneTask("c1")

	appClient.StartTask("c2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c2").WithKeyRef("m2H").WithCategory(asset.ModelCategory_MODEL_HEAD).WithTaskOutput("local"))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("c2").WithKeyRef("m2T").WithCategory(asset.ModelCategory_MODEL_SIMPLE).WithTaskOutput("shared"))
	appClient.DoneTask("c2")

	appClient.StartTask("a1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("a1").WithKeyRef("mAgg").WithCategory(asset.ModelCategory_MODEL_SIMPLE))
	appClient.DoneTask("a1")

	appClient.StartTask("c3")

	inputs := appClient.GetInputModels("c3")
	require.Equal(t, 2, len(inputs), "composite should have 2 input models")
}

func TestFailLargeComputePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	appClient := factory.NewTestClient()

	nbRounds := 1000
	nbPharma := 11
	var nbTasks int

	appClient.RegisterAlgo(client.DefaultCompositeAlgoOptions().WithKeyRef("algoComp"))
	appClient.RegisterAlgo(client.DefaultAggregateAlgoOptions().WithKeyRef("algoAgg"))
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

			task := client.DefaultCompositeTaskOptions().WithKeyRef(compKey).WithAlgoRef("algoComp")
			if i > 0 {
				// Reference previous composite and aggregate
				local := fmt.Sprintf("compP%dR%d", pharma, i-1)
				shared := fmt.Sprintf("aggR%d", i-1)
				task.
					WithParentsRef(local, shared).
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
			WithParentsRef(compKeys...).
			WithAlgoRef("algoAgg")
		for _, compKey := range compKeys {
			agg.WithInput("model", &client.TaskOutputRef{TaskRef: compKey, Identifier: "shared"})
		}

		newTasks = append(newTasks, agg)
		nbTasks++

		i++

		if i%20 == 0 {
			appClient.RegisterTasks(newTasks...)
			log.WithField("round", i).WithField("nbTasks", nbTasks).WithField("duration", time.Since(start)).Debug("Round registered")
			newTasks = make([]client.Taskable, 0)
			start = time.Now()
		}
	}

	if len(newTasks) > 0 {
		appClient.RegisterTasks(newTasks...)
		log.WithField("nbTasks", nbTasks).WithField("duration", time.Since(start)).Debug("Round registered")
	}

	// Fail the composite of rank 0 on pharma1
	start = time.Now()
	appClient.StartTask("compP1R0")
	appClient.FailTask("compP1R0")
	log.WithField("duration", time.Since(start)).WithField("nbTasks", nbTasks).Info("canceled compute plan")
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

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	// A CP with 1 parent task and 2 child tasks
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	appClient.RegisterTasks(client.DefaultTrainTaskOptions().
		WithKeyRef("task#1").
		WithParentsRef(client.DefaultTrainTaskRef).
		WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))

	plan := appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, plan.Status, asset.ComputePlanStatus_PLAN_STATUS_TODO)

	appClient.StartTask(client.DefaultTrainTaskRef)
	plan = appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, plan.Status, asset.ComputePlanStatus_PLAN_STATUS_DOING)

	appClient.FailTask("task#1")
	plan = appClient.GetComputePlan(client.DefaultPlanRef)
	require.Equal(t, plan.Status, asset.ComputePlanStatus_PLAN_STATUS_FAILED)
}

func TestCompositeParentChild(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultCompositeAlgoOptions().WithKeyRef("algoComp"))
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().WithKeyRef("comp1").WithAlgoRef("algoComp"),
	)
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("comp2").
			WithAlgoRef("algoComp").
			WithParentsRef("comp1", "comp1").
			WithInput("local", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "shared"}))

	appClient.StartTask("comp1")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1H").WithCategory(asset.ModelCategory_MODEL_HEAD).WithTaskOutput("local"))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp1").WithKeyRef("model1T").WithCategory(asset.ModelCategory_MODEL_SIMPLE).WithTaskOutput("shared"))
	appClient.DoneTask("comp1")

	appClient.StartTask("comp2")
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2H").WithCategory(asset.ModelCategory_MODEL_HEAD).WithTaskOutput("local"))
	appClient.RegisterModel(client.DefaultModelOptions().WithTaskRef("comp2").WithKeyRef("model2T").WithCategory(asset.ModelCategory_MODEL_SIMPLE).WithTaskOutput("shared"))
	appClient.DoneTask("comp2")

	// Register a composite task with 2 composite parents
	appClient.RegisterTasks(
		client.DefaultCompositeTaskOptions().
			WithKeyRef("comp3").
			WithAlgoRef("algoComp").
			WithParentsRef("comp1", "comp2").
			WithInput("local", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "local"}).
			WithInput("shared", &client.TaskOutputRef{TaskRef: "comp1", Identifier: "shared"}))

	inputs := appClient.GetInputModels("comp3")
	require.Len(t, inputs, 2, "composite task should have 2 input models")
	require.Equal(t, appClient.GetKeyStore().GetKey("model1H"), inputs[0].Key, "first model should be HEAD from comp1")
	require.Equal(t, appClient.GetKeyStore().GetKey("model2T"), inputs[1].Key, "second model should be TRUNK from comp2")
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
