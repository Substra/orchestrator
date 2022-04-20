package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var datasampleTestsScenarios = []Scenario{
	{
		testQueryDatasamplesUnfiltered,
		[]string{"short", "datasample"},
	},
	{
		testQueryDatasamplesFiltered,
		[]string{"short", "datasample"},
	},
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

	assertContainsDatasample(appClient, resp.DataSamples, "ds1")
	assertContainsDatasample(appClient, resp.DataSamples, "ds2")
	assertContainsDatasample(appClient, resp.DataSamples, "ds3")
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

	assertContainsDatasample(appClient, resp.DataSamples, "filtered_ds1")
	assertContainsDatasample(appClient, resp.DataSamples, "filtered_ds3")
}

func assertContainsDatasample(appClient *client.TestClient, datasamples []*asset.DataSample, keyRef string) {
	key := appClient.GetKeyStore().GetKey(keyRef)
	for _, ds := range datasamples {
		if ds.Key == key {
			return
		}
	}
	log.Fatal("QueryDataSamples response should contain key ref " + keyRef)
}
