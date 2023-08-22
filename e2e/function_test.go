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

// TestRegisterFunction registers an function and ensure an event containing the function is recorded.
func TestRegisterFunction(t *testing.T) {
	appClient := factory.NewTestClient()
	registeredFunction := appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())

	retrievedFunction := appClient.GetFunction(client.DefaultSimpleFunctionRef)
	e2erequire.ProtoEqual(t, registeredFunction, retrievedFunction)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredFunction.Key,
		AssetKind: asset.AssetKind_ASSET_FUNCTION,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	require.Len(t, resp.Events, 1, "Unexpected number of events")

	eventFunction := resp.Events[0].GetFunction()
	e2erequire.ProtoEqual(t, registeredFunction, eventFunction)
}

func TestQueryFunctions(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("objSample"))
	appClient.RegisterFunction(client.DefaultPredictFunctionOptions())
	appClient.RegisterFunction(client.DefaultMetricFunctionOptions())

	resp := appClient.QueryFunctions(&asset.FunctionQueryFilter{}, "", 100)

	// We cannot check for equality since this test may run after others,
	// we will probably have more than the registered function above.
	require.GreaterOrEqual(t, len(resp.Functions), 1, "Unexpected total number of function")

	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	planKey := appClient.GetKeyStore().GetKey(client.DefaultPlanRef)

	resp = appClient.QueryFunctions(&asset.FunctionQueryFilter{ComputePlanKey: planKey}, "", 100)

	require.Equal(t, 0, len(resp.Functions), "Unexpected number function used in compute plan without tasks")

	appClient.RegisterTasks(
		client.DefaultTrainTaskOptions().WithKeyRef("train1"),
		client.DefaultTrainTaskOptions().WithKeyRef("train2"),

		client.DefaultTrainTaskOptions().WithKeyRef("train3").
			WithInput("model", &client.TaskOutputRef{TaskRef: "train1", Identifier: "model"}).
			WithInput("model", &client.TaskOutputRef{TaskRef: "train2", Identifier: "model"}),

		client.DefaultPredictTaskOptions().WithKeyRef("predict").
			WithInput("model", &client.TaskOutputRef{TaskRef: "train3", Identifier: "model"}),

		client.DefaultTestTaskOptions().WithDataSampleRef("objSample").
			WithInput("predictions", &client.TaskOutputRef{TaskRef: "predict", Identifier: "predictions"}),
	)

	resp = appClient.QueryFunctions(&asset.FunctionQueryFilter{ComputePlanKey: planKey}, "", 100)
	require.Equal(t, 3, len(resp.Functions), "Unexpected number of function used in compute plan with tasks")
}

func TestQueryFunctionsInputOutputs(t *testing.T) {
	appClient := factory.NewTestClient()

	keyRef := "test-functions-input-outputs"
	key := appClient.GetKeyStore().GetKey(keyRef)

	functionOptions := client.DefaultSimpleFunctionOptions().WithKeyRef(keyRef)
	functionOptions.Inputs = map[string]*asset.FunctionInput{
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
	functionOptions.Outputs = map[string]*asset.FunctionOutput{
		"model": {
			Kind:     asset.AssetKind_ASSET_MODEL,
			Multiple: true,
		},
		"performance": {
			Kind: asset.AssetKind_ASSET_PERFORMANCE,
		},
	}
	appClient.RegisterFunction(functionOptions)

	// test QueryFunctions
	resp := appClient.QueryFunctions(nil, "", 10000)
	found := false
	for _, function := range resp.Functions {
		if function.Key == key {
			found = true
			e2erequire.ProtoMapEqual(t, function.Inputs, functionOptions.Inputs)
			e2erequire.ProtoMapEqual(t, function.Outputs, functionOptions.Outputs)
			break
		}
	}
	require.True(t, found, "Could not find expected function with key ref "+keyRef)

	// test GetFunction
	respFunction := appClient.GetFunction(keyRef)

	e2erequire.ProtoMapEqual(t, respFunction.Inputs, functionOptions.Inputs)
	e2erequire.ProtoMapEqual(t, respFunction.Outputs, functionOptions.Outputs)
}

// TestUpdateFunction updates mutable fieds of an function and ensure an event containing the function is recorded. List of mutable fields: name.
func TestUpdateFunction(t *testing.T) {
	appClient := factory.NewTestClient()
	keyRef := "function_filter_simple"
	registeredFunction := appClient.RegisterFunction(client.DefaultSimpleFunctionOptions().WithKeyRef(keyRef))

	appClient.UpdateFunction(keyRef, "new function name")

	expectedFunction := registeredFunction
	expectedFunction.Name = "new function name"

	retrievedFunction := appClient.GetFunction(keyRef)
	e2erequire.ProtoEqual(t, expectedFunction, retrievedFunction)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredFunction.Key,
		AssetKind: asset.AssetKind_ASSET_FUNCTION,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
	}, "", 100)

	require.Len(t, resp.Events, 1, "Unexpected number of events")

	eventFunction := resp.Events[0].GetFunction()
	e2erequire.ProtoEqual(t, expectedFunction, eventFunction)
}

func TestUpdateFunctionStatusBuilding(t *testing.T) {
	appClient := factory.NewTestClient()
	keyRef := "function_filter_simple"
	registeredFunction := appClient.RegisterFunction(client.DefaultSimpleFunctionOptions().WithKeyRef(keyRef))
	status := asset.FunctionStatus_FUNCTION_STATUS_BUILDING
	appClient.UpdateFunctionStatus(keyRef, status)

	expectedFunction := registeredFunction
	expectedFunction.Status = status

	retrievedFunction := appClient.GetFunction(keyRef)
	e2erequire.ProtoEqual(t, expectedFunction, retrievedFunction)

	resp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredFunction.Key,
		AssetKind: asset.AssetKind_ASSET_FUNCTION,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
	}, "", 100)

	require.Len(t, resp.Events, 1, "Unexpected number of events")

	eventFunction := resp.Events[0].GetFunction()
	e2erequire.ProtoEqual(t, expectedFunction, eventFunction)
}
