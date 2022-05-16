package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/golang/protobuf/proto"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var datasampleTestsScenarios = []Scenario{
	{
		testRegisterDataSample,
		[]string{"short", "datasample"},
	},
	{
		testQueryDatasamplesUnfiltered,
		[]string{"short", "datasample"},
	},
	{
		testQueryDatasamplesFiltered,
		[]string{"short", "datasample"},
	},
}

func testRegisterDataSample(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	registeredSample := appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	retrievedSample := appClient.GetDataSample(client.DefaultDataSampleRef)

	if !proto.Equal(registeredSample, retrievedSample) {
		log.WithField("registeredSample", registeredSample).WithField("retrievedSample", retrievedSample).
			Fatal("The retrieved datasample differs from the registered datasample")
	}
}

func testQueryDatasamplesUnfiltered(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds2"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds3"))

	resp := appClient.QueryDataSamples("", 1000, nil)

	if len(resp.DataSamples) < 3 {
		log.Fatal("QueryDataSamples response should contain at least 3 datasamples")
	}

	assertContainsKeys(true, appClient, resp.DataSamples, "ds1", "ds2", "ds3")
}

func testQueryDatasamplesFiltered(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("filtered_ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("filtered_ds2"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("filtered_ds3"))

	keyStore := appClient.GetKeyStore()
	targetKeys := []string{keyStore.GetKey("filtered_ds1"), keyStore.GetKey("filtered_ds3")}

	resp := appClient.QueryDataSamples("", 10, &asset.DataSampleQueryFilter{Keys: targetKeys})

	if len(resp.DataSamples) != 2 {
		log.Fatal("QueryDataSamples response should contain 2 datasamples")
	}

	assertContainsKeys(true, appClient, resp.DataSamples, "filtered_ds1", "filtered_ds3")
}
