// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestDataSampleAdapterImplementServer(t *testing.T) {
	adapter := NewDataSampleAdapter()
	assert.Implementsf(t, (*asset.DataSampleServiceServer)(nil), adapter, "DataSampleAdapter should implements DataSampleServiceServer")
}

func TestRegisterDataSample(t *testing.T) {
	adapter := NewDataSampleAdapter()

	newDS := &asset.NewDataSample{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		TestOnly:        false,
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.datasample:RegisterDataSample", newDS, nil).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterDataSample(ctx, newDS)

	assert.NoError(t, err, "Registration should pass")
}

func TestUpdateDataSample(t *testing.T) {
	adapter := NewDataSampleAdapter()

	updatedDS := &asset.DataSampleUpdateParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.datasample:UpdateDataSample", updatedDS, nil).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.UpdateDataSample(ctx, updatedDS)
	assert.NoError(t, err, "Update should pass")
}

func TestQueryDataSamples(t *testing.T) {
	adapter := NewDataSampleAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	queryParam := &asset.DataSamplesQueryParam{PageToken: "", PageSize: 10}
	invocator.On("Call", "orchestrator.datasample:QueryDataSamples", queryParam, &asset.DataSamplesQueryResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryDataSamples(ctx, queryParam)

	assert.NoError(t, err, "Query should pass")
}
