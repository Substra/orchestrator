//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
)

func TestRegisterDataManager(t *testing.T) {
	appClient := factory.NewTestClient()

	registeredManager := appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	retrievedManager := appClient.GetDataManager(client.DefaultDataManagerRef)

	e2erequire.ProtoEqual(t, registeredManager, retrievedManager)
}

// TestUpdateDataManager updates mutable fieds of a data manager and ensure an event containing the data manager is recorded. List of mutable fields: name.
func TestUpdateDataManager(t *testing.T) {
	appClient := factory.NewTestClient()
	registeredDataManager := appClient.RegisterDataManager(client.DefaultDataManagerOptions())

	appClient.UpdateDataManager(client.DefaultDataManagerRef, "new data manager name")

	expectedDataManager := registeredDataManager
	expectedDataManager.Name = "new data manager name"

	retrievedDataManager := appClient.GetDataManager(client.DefaultDataManagerRef)
	e2erequire.ProtoEqual(t, expectedDataManager, retrievedDataManager)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredDataManager.Key,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
	}, "", 100)

	require.Len(t, resp.Events, 1, "Unexpected number of events")

	eventDataManager := resp.Events[0].GetDataManager()
	e2erequire.ProtoEqual(t, expectedDataManager, eventDataManager)
}
