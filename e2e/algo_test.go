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

// TestRegisterAlgo registers an algo and ensure an event containing the algo is recorded.
func TestRegisterAlgo(t *testing.T) {
	appClient := factory.NewTestClient()
	registeredAlgo := appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())

	retrievedAlgo := appClient.GetAlgo(client.DefaultSimpleAlgoRef)
	e2erequire.ProtoEqual(t, registeredAlgo, retrievedAlgo)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredAlgo.Key,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	require.Len(t, resp.Events, 1, "Unexpected number of events")

	eventAlgo := resp.Events[0].GetAlgo()
	e2erequire.ProtoEqual(t, registeredAlgo, eventAlgo)
}

func TestQueryAlgos(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample").WithTestOnly(true))
	appClient.RegisterAlgo(client.DefaultPredictAlgoOptions())
	appClient.RegisterAlgo(client.DefaultMetricAlgoOptions())

	resp := appClient.QueryAlgos(&asset.AlgoQueryFilter{}, "", 100)

	// We cannot check for equality since this test may run after others,
	// we will probably have more than the registered algo above.
	require.GreaterOrEqual(t, len(resp.Algos), 1, "Unexpected total number of algo")

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	planKey := appClient.GetKeyStore().GetKey(client.DefaultPlanRef)

	resp = appClient.QueryAlgos(&asset.AlgoQueryFilter{ComputePlanKey: planKey}, "", 100)

	require.Equal(t, 0, len(resp.Algos), "Unexpected number algo used in compute plan without tasks")

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),
		client.DefaultTrainTaskOptions().WithKeyRef("train3").WithParentsRef("train1", "train2"),
		client.DefaultPredictTaskOptions().WithKeyRef("predict").WithParentsRef("train3"),
		client.DefaultTestTaskOptions().WithDataSampleRef("objSample").WithParentsRef("predict"),
	)

	resp = appClient.QueryAlgos(&asset.AlgoQueryFilter{ComputePlanKey: planKey}, "", 100)
	require.Equal(t, 3, len(resp.Algos), "Unexpected number of algo used in compute plan with tasks")
}

func TestQueryAlgosInputOutputs(t *testing.T) {
	appClient := factory.NewTestClient()

	keyRef := "test-algos-input-outputs"
	key := appClient.GetKeyStore().GetKey(keyRef)

	algoOptions := client.DefaultSimpleAlgoOptions().WithKeyRef(keyRef)
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
			e2erequire.ProtoMapEqual(t, algo.Inputs, algoOptions.Inputs)
			e2erequire.ProtoMapEqual(t, algo.Outputs, algoOptions.Outputs)
			break
		}
	}
	require.True(t, found, "Could not find expected algo with key ref "+keyRef)

	// test GetAlgo
	respAlgo := appClient.GetAlgo(keyRef)

	e2erequire.ProtoMapEqual(t, respAlgo.Inputs, algoOptions.Inputs)
	e2erequire.ProtoMapEqual(t, respAlgo.Outputs, algoOptions.Outputs)
}
