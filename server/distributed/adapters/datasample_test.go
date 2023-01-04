package adapters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/distributed/chaincode"
	"github.com/substra/orchestrator/server/distributed/interceptors"
	"github.com/substra/orchestrator/utils"
)

func TestDataSampleAdapterImplementServer(t *testing.T) {
	adapter := NewDataSampleAdapter()
	assert.Implementsf(t, (*asset.DataSampleServiceServer)(nil), adapter, "DataSampleAdapter should implements DataSampleServiceServer")
}

func TestRegisterDataSample(t *testing.T) {
	adapter := NewDataSampleAdapter()

	param := &asset.RegisterDataSamplesParam{
		Samples: []*asset.NewDataSample{
			{
				Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
				DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			},
		},
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datasample:RegisterDataSamples", param, &asset.RegisterDataSamplesResponse{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterDataSamples(ctx, param)

	assert.NoError(t, err, "Registration should pass")
}

func TestUpdateDataSamples(t *testing.T) {
	adapter := NewDataSampleAdapter()

	updatedDS := &asset.UpdateDataSamplesParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datasample:UpdateDataSamples", updatedDS, nil).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.UpdateDataSamples(ctx, updatedDS)
	assert.NoError(t, err, "Update should pass")
}

func TestQueryDataSamples(t *testing.T) {
	adapter := NewDataSampleAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	queryParam := &asset.QueryDataSamplesParam{PageToken: "", PageSize: 10}
	invocator.On("Call", utils.AnyContext, "orchestrator.datasample:QueryDataSamples", queryParam, &asset.QueryDataSamplesResponse{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.QueryDataSamples(ctx, queryParam)

	assert.NoError(t, err, "Query should pass")
}

func TestHandleDataSampleConflictAfterTimeout(t *testing.T) {
	adapter := NewDataSampleAdapter()

	param := &asset.RegisterDataSamplesParam{
		Samples: []*asset.NewDataSample{
			{
				Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
				DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			},
		},
	}

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datasample:RegisterDataSamples", param, &asset.RegisterDataSamplesResponse{}).
		Return(errors.NewError(errors.ErrConflict, "test"))

	invocator.On("Call", utils.AnyContext, "orchestrator.datasample:GetDataSample", &asset.GetDataSampleParam{Key: "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"}, &asset.DataSample{}).
		Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterDataSamples(ctx, param)

	assert.NoError(t, err, "Registration should pass")
}

func TestHandleDataSampleBatchConflictAfterTimeout(t *testing.T) {
	adapter := NewDataSampleAdapter()

	param := &asset.RegisterDataSamplesParam{
		Samples: []*asset.NewDataSample{
			{
				Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
				DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			},
			{
				Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84",
				DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			},
		},
	}

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datasample:RegisterDataSamples", param, &asset.RegisterDataSamplesResponse{}).
		Return(errors.NewError(errors.ErrConflict, "test"))

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterDataSamples(ctx, param)

	// We cannot assume that ALL the assets have been created, it might be a legitimate conflict not due to the timeout.
	assert.Error(t, err, "Registration fail because batch contains more than one sample")
}

func TestGetDataSample(t *testing.T) {
	adapter := NewDataSampleAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.GetDataSampleParam{Key: "uuid"}

	invocator.On("Call", utils.AnyContext, "orchestrator.datasample:GetDataSample", param, &asset.DataSample{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.GetDataSample(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
