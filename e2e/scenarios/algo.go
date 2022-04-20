package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/golang/protobuf/proto"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var algoTestScenarios = []Scenario{
	{
		testQueryAlgos,
		[]string{"short", "algo"},
	},
	{
		testQueryAlgosFilterCategories,
		[]string{"short", "algo"},
	},
	{
		testQueryAlgosInputOutputs,
		[]string{"short", "algo"},
	},
}

func testQueryAlgos(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithKeyRef(client.DefaultMetricRef).WithCategory(asset.AlgoCategory_ALGO_METRIC))

	resp := appClient.QueryAlgos(&asset.AlgoQueryFilter{}, "", 100)

	// We cannot check for equality since this test may run after others,
	// we will probably have more than the registered algo above.
	if len(resp.Algos) < 1 {
		log.WithField("numAlgos", len(resp.Algos)).Fatal("Unexpected total number of algo")
	}

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	planKey := appClient.GetKeyStore().GetKey(client.DefaultPlanRef)

	resp = appClient.QueryAlgos(&asset.AlgoQueryFilter{ComputePlanKey: planKey}, "", 100)
	if len(resp.Algos) != 0 {
		log.WithField("numAlgos", len(resp.Algos)).Fatal("Unexpected number algo used in compute plan without tasks")
	}

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),
		client.DefaultTrainTaskOptions().WithKeyRef("train3").WithParentsRef("train1", "train2"),
		client.DefaultTestTaskOptions().WithDataSampleRef("objSample").WithParentsRef("train3"),
	)

	resp = appClient.QueryAlgos(&asset.AlgoQueryFilter{ComputePlanKey: planKey}, "", 100)
	if len(resp.Algos) != 1 {
		log.WithField("numAlgos", len(resp.Algos)).Fatal("Unexpected number of algo used in compute plan with tasks")
	}
}

func testQueryAlgosFilterCategories(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_SIMPLE).WithKeyRef("algo_filter_simple"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_COMPOSITE).WithKeyRef("algo_filter_composite"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_AGGREGATE).WithKeyRef("algo_filter_aggregate"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("algo_filter_metric"))

	resp := appClient.QueryAlgos(&asset.AlgoQueryFilter{}, "", 10000)

	if len(resp.Algos) < 4 {
		log.WithField("numAlgos", len(resp.Algos)).Fatal("Unexpected number of algos")
	}

	assertContainsAlgo(true, appClient, resp.Algos, "algo_filter_simple")
	assertContainsAlgo(true, appClient, resp.Algos, "algo_filter_composite")
	assertContainsAlgo(true, appClient, resp.Algos, "algo_filter_aggregate")
	assertContainsAlgo(true, appClient, resp.Algos, "algo_filter_metric")

	filter := &asset.AlgoQueryFilter{
		Categories: []asset.AlgoCategory{
			asset.AlgoCategory_ALGO_SIMPLE,
			asset.AlgoCategory_ALGO_METRIC,
		}}

	resp = appClient.QueryAlgos(filter, "", 100)

	assertContainsAlgo(true, appClient, resp.Algos, "algo_filter_simple")
	assertContainsAlgo(false, appClient, resp.Algos, "algo_filter_composite")
	assertContainsAlgo(false, appClient, resp.Algos, "algo_filter_aggregate")
	assertContainsAlgo(true, appClient, resp.Algos, "algo_filter_metric")
}

func testQueryAlgosInputOutputs(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	keyRef := "test-algos-input-outputs"
	key := appClient.GetKeyStore().GetKey(keyRef)

	algoOptions := client.DefaultAlgoOptions().WithKeyRef(keyRef)
	algoOptions.Inputs = map[string]*asset.AlgoInput{
		"data manager": {
			Kind: asset.AssetKind_ASSET_DATA_MANAGER,
		},
		"data samples": {
			Kind:     asset.AssetKind_ASSET_DATA_SAMPLE,
			Multiple: true,
		},
		"model": {
			Kind:     asset.AssetKind_ASSET_MODEL,
			Optional: true,
		},
	}
	algoOptions.Outputs = map[string]*asset.AlgoOutput{
		"model": {
			Kind:     asset.AssetKind_ASSET_MODEL,
			Multiple: true,
		},
		"performance": {
			Kind: asset.AssetKind_ASSET_PERFORMANCE,
		},
	}
	appClient.RegisterAlgo(algoOptions)

	// test QueryAlgos
	resp := appClient.QueryAlgos(nil, "", 10000)
	found := false
	for _, algo := range resp.Algos {
		if algo.Key == key {
			found = true
			if !ResourcesEqual(algo.Inputs, algoOptions.Inputs) {
				log.WithField("actual", algo.Inputs).WithField("expected", algoOptions.Inputs).Fatal("Unexpected algo inputs")
			}
			if !ResourcesEqual(algo.Outputs, algoOptions.Outputs) {
				log.WithField("actual", algo.Outputs).WithField("expected", algoOptions.Outputs).Fatal("Unexpected algo outputs")
			}
		}
	}
	if !found {
		log.Fatal("Could not find expected algo with key ref " + keyRef)
	}

	// test GetAlgo
	respAlgo := appClient.GetAlgo(keyRef)

	if !ResourcesEqual(respAlgo.Inputs, algoOptions.Inputs) {
		log.WithField("actual", respAlgo.Inputs).WithField("expected", algoOptions.Inputs).Fatal("Unexpected algo inputs")
	}
	if !ResourcesEqual(respAlgo.Outputs, algoOptions.Outputs) {
		log.WithField("actual", respAlgo.Outputs).WithField("expected", algoOptions.Outputs).Fatal("Unexpected algo outputs")
	}
}

func ResourcesEqual[T proto.Message](a, b map[string]T) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v2, ok := b[k]; !ok || !proto.Equal(v, v2) {
			return false
		}
	}
	return true
}

func assertContainsAlgo(shouldContain bool, appClient *client.TestClient, algos []*asset.Algo, keyRef string) {
	key := appClient.GetKeyStore().GetKey(keyRef)
	for _, ds := range algos {
		if ds.Key == key {
			if shouldContain {
				return
			}
			log.Fatal("QueryAlgos response should NOT contain key ref " + keyRef)
		}
	}
	if shouldContain {
		log.Fatal("QueryAlgos response should contain key ref " + keyRef)
	}
}
