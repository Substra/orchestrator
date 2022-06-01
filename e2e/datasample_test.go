//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/owkin/orchestrator/e2e/client"
	e2erequire "github.com/owkin/orchestrator/e2e/require"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/require"
)

// TestRegisterDataSample registers a datasample and ensure an event containing the datasample is recorded.
func TestRegisterDataSample(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	registeredSample := appClient.RegisterDataSample(client.DefaultDataSampleOptions())

	retrievedSample := appClient.GetDataSample(client.DefaultDataSampleRef)
	e2erequire.ProtoEqual(t, registeredSample, retrievedSample)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredSample.Key,
		AssetKind: asset.AssetKind_ASSET_DATA_SAMPLE,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	require.Len(t, resp.Events, 1)

	eventSample := resp.Events[0].GetDataSample()
	e2erequire.ProtoEqual(t, registeredSample, eventSample)
}

func TestQueryDatasamplesUnfiltered(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds2"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds3"))

	resp := appClient.QueryDataSamples("", 1000, nil)

	require.Greater(t, len(resp.DataSamples), 3, "QueryDataSamples response should contain at least 3 datasamples")
	e2erequire.ContainsKeys(t, true, appClient, resp.DataSamples, "ds1", "ds2", "ds3")
}

func TestQueryDatasamplesFiltered(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("filtered_ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("filtered_ds2"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("filtered_ds3"))

	keyStore := appClient.GetKeyStore()
	targetKeys := []string{keyStore.GetKey("filtered_ds1"), keyStore.GetKey("filtered_ds3")}

	resp := appClient.QueryDataSamples("", 10, &asset.DataSampleQueryFilter{Keys: targetKeys})

	require.Equal(t, 2, len(resp.DataSamples), "QueryDataSamples response should contain 2 datasamples")
	e2erequire.ContainsKeys(t, true, appClient, resp.DataSamples, "filtered_ds1", "filtered_ds3")
}
