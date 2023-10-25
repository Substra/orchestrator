package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalEventAsset(t *testing.T) {
	cases := map[string]*Event{
		"function": {
			AssetKind: AssetKind_ASSET_FUNCTION,
			Asset:     &Event_Function{Function: &Function{Key: "function"}},
		},
		"computePlan": {
			AssetKind: AssetKind_ASSET_COMPUTE_PLAN,
			Asset:     &Event_ComputePlan{ComputePlan: &ComputePlan{Key: "plan"}},
		},
		"computeTask": {
			AssetKind: AssetKind_ASSET_COMPUTE_TASK,
			Asset:     &Event_ComputeTask{ComputeTask: &ComputeTask{Key: "task"}},
		},
		"dataManager": {
			AssetKind: AssetKind_ASSET_DATA_MANAGER,
			Asset:     &Event_DataManager{DataManager: &DataManager{Key: "manager"}},
		},
		"dataSample": {
			AssetKind: AssetKind_ASSET_DATA_SAMPLE,
			Asset:     &Event_DataSample{DataSample: &DataSample{Key: "sample"}},
		},
		"failureReport": {
			AssetKind: AssetKind_ASSET_FAILURE_REPORT,
			Asset:     &Event_FailureReport{FailureReport: &FailureReport{AssetKey: "failed-task"}},
		},
		"model": {
			AssetKind: AssetKind_ASSET_MODEL,
			Asset:     &Event_Model{Model: &Model{Key: "model"}},
		},
		"organization": {
			AssetKind: AssetKind_ASSET_ORGANIZATION,
			Asset:     &Event_Organization{Organization: &Organization{Id: "organization"}},
		},
		"performance": {
			AssetKind: AssetKind_ASSET_PERFORMANCE,
			Asset:     &Event_Performance{Performance: &Performance{ComputeTaskKey: "test-task"}},
		},
	}

	for name, event := range cases {
		marshalled, err := MarshalEventAsset(event)
		require.NoError(t, err, name+" marshalling has failed")

		unmarshalled := &Event{AssetKind: event.AssetKind}
		err = UnmarshalEventAsset(marshalled, unmarshalled, unmarshalled.AssetKind)
		require.NoError(t, err, name+" unmarshalling has failed")

		assert.Equal(t, event, unmarshalled)
	}
}

func TestMarshalEventAssetErrorNilAsset(t *testing.T) {
	_, err := MarshalEventAsset(&Event{Asset: nil})
	assert.ErrorContains(t, err, "unsupported asset")
}

func TestUnmarshalEventAssetErrorUnsupportedAssetKind(t *testing.T) {
	err := UnmarshalEventAsset([]byte{}, &Event{}, AssetKind_ASSET_UNKNOWN)
	assert.ErrorContains(t, err, "unsupported asset kind")
}
