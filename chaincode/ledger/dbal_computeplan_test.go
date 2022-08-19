package ledger

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/utils"
)

func TestQueryComputePlans(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxIsEvaluateTransaction, true)
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(ctx, stub, queue)

	// ComputePlan iterator
	resp := &testHelper.MockedStateQueryIterator{}
	resp.On("Close").Return(nil)
	resp.On("HasNext").Once().Return(true)
	resp.On("HasNext").Once().Return(false)
	resp.On("Next").Once().Return(&queryresult.KV{Value: []byte(`{"asset":{"key":"uuid"}}`)}, nil)

	queryString := `{"selector":{"doc_type":"computeplan","asset":{"owner":"owner"}}}`
	stub.On("GetQueryResultWithPagination", queryString, int32(1), "").
		Return(resp, &peer.QueryResponseMetadata{Bookmark: "", FetchedRecordsCount: 1}, nil)

	// CP tasks index
	index := &testHelper.MockedStateQueryIterator{}
	index.On("Close").Return(nil)
	index.On("HasNext").Twice().Return(true)
	index.On("HasNext").Once().Return(false)
	index.On("Next").Return(&queryresult.KV{Key: "firstIndexKey"}, nil)

	stub.On("SplitCompositeKey", "firstIndexKey").Return("", []string{"indexName", "uuid", "STATUS_FAILED", "taskId"}, nil)

	stub.On("GetStateByPartialCompositeKey", computePlanTaskStatusIndex, []string{asset.ComputePlanKind, "uuid"}).
		Return(index, nil)

	_, _, err := db.QueryComputePlans(
		common.NewPagination("", 1),
		&asset.PlanQueryFilter{Owner: "owner"},
	)
	assert.NoError(t, err)

	stub.AssertExpectations(t)
}

func TestQueryComputePlansNilFilter(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxIsEvaluateTransaction, true)
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(ctx, stub, queue)

	// ComputePlan iterator
	resp := &testHelper.MockedStateQueryIterator{}
	resp.On("Close").Return(nil)
	resp.On("HasNext").Once().Return(false)

	queryString := `{"selector":{"doc_type":"computeplan"}}`
	stub.On("GetQueryResultWithPagination", queryString, int32(1), "").
		Return(resp, &peer.QueryResponseMetadata{Bookmark: "", FetchedRecordsCount: 0}, nil)

	_, _, err := db.QueryComputePlans(
		common.NewPagination("", 1),
		nil,
	)
	assert.NoError(t, err)

	stub.AssertExpectations(t)
}

// TestStorableComputePlanFields should fail if the storable struct is not updated after a new field is added to the compute plan.
func TestStorableComputePlanFields(t *testing.T) {
	var publicPlanFields, publicStorablePlanFields int

	// Represents the list of compute plan dynamically computed fields.
	// planComputedFields should be updated after adding a new dynamic field to the compute plan.
	planComputedFields := []string{
		"WaitingCount",
		"TodoCount",
		"DoingCount",
		"CanceledCount",
		"FailedCount",
		"DoneCount",
		"TaskCount",
		"Status",
	}

	planType := reflect.TypeOf(asset.ComputePlan{})
	planFields := reflect.VisibleFields(planType)
	for _, f := range planFields {
		if f.IsExported() && !utils.SliceContains(planComputedFields, f.Name) {
			publicPlanFields++
		}
	}

	storableType := reflect.TypeOf(storableComputePlan{})
	storableFields := reflect.VisibleFields(storableType)
	for _, f := range storableFields {
		if f.IsExported() {
			publicStorablePlanFields++
		}
	}

	assert.GreaterOrEqual(t, publicStorablePlanFields, publicPlanFields, "storable struct should have at least as many fields than the asset it represents")
}

func TestStorableComputeTaskOutputAsset(t *testing.T) {
	output := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              "uuid",
		ComputeTaskOutputIdentifier: "test",
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    "testasset",
	}

	storable := newStorableTaskOutputAsset(output)

	converted, err := storable.newComputeTaskOutputAsset()
	require.NoError(t, err)
	assert.Equal(t, output, converted)
}

func TestGetComputeTaskOutputAssets(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxIsEvaluateTransaction, true)
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(ctx, stub, queue)

	storableOutputs := []*storableComputeTaskOutputAsset{
		{ComputeTaskKey: "uuid", ComputeTaskOutputIdentifier: "model", AssetKind: asset.AssetKind_ASSET_MODEL.String(), AssetKey: "modelUUID"},
		{ComputeTaskKey: "uuid", ComputeTaskOutputIdentifier: "anotherIdentifier", AssetKind: asset.AssetKind_ASSET_MODEL.String(), AssetKey: "anotherModelUUID"},
	}
	bytes, err := json.Marshal(storableOutputs)
	require.NoError(t, err)
	stored := storedAsset{
		DocType: "test",
		Asset:   bytes,
	}
	bytes, err = json.Marshal(stored)
	require.NoError(t, err)

	stub.On("GetState", "computetask_output_asset:uuid").
		Twice(). // First for hasKey, then to get the actual value
		Return(bytes, nil)

	outputs, err := db.GetComputeTaskOutputAssets("uuid", "model")
	require.NoError(t, err)

	expectedOutput := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              "uuid",
		ComputeTaskOutputIdentifier: "model",
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    "modelUUID",
	}

	assert.Len(t, outputs, 1)
	assert.Equal(t, expectedOutput, outputs[0])

	stub.AssertExpectations(t)
}
