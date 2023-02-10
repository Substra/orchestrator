//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
)

func TestDatasetSampleKeys(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds1"))
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("ds2"))

	dataset := appClient.GetDataset(client.DefaultDataManagerRef)

	require.Equal(t, 2, len(dataset.DataSampleKeys), "dataset should contain 2 data samples")
	require.Equal(t, appClient.GetKeyStore().GetKey("ds1"), dataset.DataSampleKeys[0], "dataset should contain valid data sample ID")
}
