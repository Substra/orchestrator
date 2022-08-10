//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/substra/orchestrator/e2e/client"
	"github.com/stretchr/testify/require"
)

func TestDatasetSampleKeys(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds2"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithTestOnly(true).WithKeyRef("testds"))

	dataset := appClient.GetDataset(client.DefaultDataManagerRef)

	require.Equal(t, 1, len(dataset.TestDataSampleKeys), "dataset should contain a single test sample")
	require.Equal(t, 2, len(dataset.TrainDataSampleKeys), "dataset should contain 2 train samples")
	require.Equal(t, appClient.GetKeyStore().GetKey("testds"), dataset.TestDataSampleKeys[0], "dataset should contain valid test sample ID")
}
