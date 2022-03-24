package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
)

var datasetTestScenarios = []Scenario{
	{
		testDatasetSampleKeys,
		[]string{"short", "dataset"},
	},
}

func testDatasetSampleKeys(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds2"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithTestOnly(true).WithKeyRef("testds"))

	dataset := appClient.GetDataset(client.DefaultDataManagerRef)

	if len(dataset.TestDataSampleKeys) != 1 {
		log.Fatal("dataset should contain a single test sample")
	}
	if len(dataset.TrainDataSampleKeys) != 2 {
		log.Fatal("dataset should contain 2 train samples")
	}
	if dataset.TestDataSampleKeys[0] != appClient.GetKeyStore().GetKey("testds") {
		log.Fatal("dataset should contain valid test sample ID")
	}
}
