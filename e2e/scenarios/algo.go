package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var algoTestScenarios = []Scenario{
	{
		testQueryAlgos,
		[]string{"short", "algo"},
	},
}

func testQueryAlgos(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterMetric(client.DefaultMetricOptions())

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
