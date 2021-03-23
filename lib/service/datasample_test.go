// Copyright 2020 Owkin Inc.
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

package service

import (
	"errors"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterSingleDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	datasample := &asset.NewDataSample{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		TestOnly:        false,
	}

	storedDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	dbal.On("AddDataSample", storedDataSample).Return(nil).Once()
	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(nil).Once()

	err := service.RegisterDataSample(datasample, "owner")

	assert.NoError(t, err, "Registration of valid datasample should not fail")

	dbal.AssertExpectations(t)
}

func TestRegisterSingleDataSampleUnknownDataManager(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	datasample := &asset.NewDataSample{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		TestOnly:        false,
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(errors.New("unknown datamanager")).Once()

	err := service.RegisterDataSample(datasample, "owner")

	assert.Error(t, err, "Registration of datasample with invalid datamanager key should fail")

	dbal.AssertExpectations(t)
}

func TestRegisterMultipleDataSamples(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	datasamples := &asset.NewDataSample{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "0b4b4466-9a81-4084-9bab-80939b78addd"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		TestOnly:        false,
	}

	storedDataSample1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	storedDataSample2 := &asset.DataSample{
		Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(nil).Once()
	dbal.On("AddDataSample", storedDataSample1).Return(nil).Once()
	dbal.On("AddDataSample", storedDataSample2).Return(nil).Once()

	err := service.RegisterDataSample(datasamples, "owner")

	assert.NoError(t, err, "Registration of multiple valid assets should not fail")

	dbal.AssertExpectations(t)
}

func TestUpdateSingleExistingDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	existingDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	updatedDataSample := &asset.DataSampleUpdateParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
	}

	storedDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
		TestOnly:        false,
	}

	dbal.On("GetDataSample", existingDataSample.GetKey()).Return(existingDataSample, nil).Once()
	dbal.On("UpdateDataSample", storedDataSample).Return(nil).Once()
	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"}, "owner").Return(nil).Once()

	err := service.UpdateDataSample(updatedDataSample, "owner")

	assert.NoError(t, err, "Update of single valid assets should not fail")

	dbal.AssertExpectations(t)
}

func TestUpdateMultipleExistingDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	existingDataSample1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	existingDataSample2 := &asset.DataSample{
		Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	updatedDataSample := &asset.DataSampleUpdateParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "0b4b4466-9a81-4084-9bab-80939b78addd"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
	}

	storedDataSample1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
		TestOnly:        false,
	}

	storedDataSample2 := &asset.DataSample{
		Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
		TestOnly:        false,
	}

	dbal.On("GetDataSample", existingDataSample1.GetKey()).Return(existingDataSample1, nil).Once()
	dbal.On("GetDataSample", existingDataSample2.GetKey()).Return(existingDataSample2, nil).Once()
	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"}, "owner").Return(nil).Once()
	dbal.On("UpdateDataSample", storedDataSample1).Return(nil).Once()
	dbal.On("UpdateDataSample", storedDataSample2).Return(nil).Once()

	err := service.UpdateDataSample(updatedDataSample, "owner")

	assert.NoError(t, err, "Update of single valid assets should not fail")

	dbal.AssertExpectations(t)
}

func TestUpdateSingleNewDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	updatedDataSample := &asset.DataSampleUpdateParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"}, "owner").Return(nil).Once()
	dbal.On("GetDataSample", updatedDataSample.GetKeys()[0]).Return(&asset.DataSample{}, errors.New("sql Error")).Once()

	err := service.UpdateDataSample(updatedDataSample, "owner")

	assert.Error(t, err, "Update of single unknown asset should fail")

	dbal.AssertExpectations(t)
}

func TestGetDataSamples(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	service := NewDataSampleService(provider)

	ds1 := asset.DataSample{
		Key:      "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		TestOnly: true,
	}
	ds2 := asset.DataSample{
		Key:      "9eef1e88-951a-44fb-944a-c3dbd1d72d85",
		TestOnly: true,
	}

	pagination := common.NewPagination("", 10)

	dbal.On("GetDataSamples", pagination).Return([]*asset.DataSample{&ds1, &ds2}, "nextPage", nil).Once()

	r, token, err := service.GetDataSamples(pagination)

	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, ds1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}
