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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterSingleDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	es := new(MockEventService)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	datasample := &asset.NewDataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		TestOnly:        false,
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	storedDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	dbal.On("DataSampleExists", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83").Return(false, nil).Once()
	dbal.On("AddDataSample", storedDataSample).Return(nil).Once()
	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_DATA_SAMPLE,
		AssetKey:  storedDataSample.Key,
	}
	es.On("RegisterEvent", e).Once().Return(nil)

	err := service.registerDataSample(datasample, "owner")

	assert.NoError(t, err, "Registration of valid datasample should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestRegisterSingleDataSampleUnknownDataManager(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	datasample := &asset.NewDataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		TestOnly:        false,
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(errors.New("unknown datamanager")).Once()

	err := service.registerDataSample(datasample, "owner")

	assert.Error(t, err, "Registration of datasample with invalid datamanager key should fail")

	dbal.AssertExpectations(t)
}

func TestRegisterMultipleDataSamples(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	es := new(MockEventService)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	datasamples := []*asset.NewDataSample{
		{
			Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			TestOnly:        false,
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
		{
			Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
			DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			TestOnly:        false,
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	storedDataSample1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	storedDataSample2 := &asset.DataSample{
		Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(nil).Times(2)
	dbal.On("DataSampleExists", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83").Return(false, nil).Once()
	dbal.On("DataSampleExists", "0b4b4466-9a81-4084-9bab-80939b78addd").Return(false, nil).Once()
	dbal.On("AddDataSample", storedDataSample1).Return(nil).Once()
	dbal.On("AddDataSample", storedDataSample2).Return(nil).Once()

	es.On("RegisterEvent", mock.AnythingOfType("*asset.Event")).Times(2).Return(nil)

	err := service.RegisterDataSamples(datasamples, "owner")

	assert.NoError(t, err, "Registration of multiple valid assets should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateSingleExistingDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	es := new(MockEventService)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	existingDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	updatedDataSample := &asset.UpdateDataSamplesParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"4da124eb-4da3-45e2-bc61-1924be259032"},
	}

	storedDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
		TestOnly:        false,
	}

	dbal.On("GetDataSample", existingDataSample.GetKey()).Return(existingDataSample, nil).Once()
	dbal.On("UpdateDataSample", storedDataSample).Return(nil).Once()
	dm.On("CheckOwner", []string{"4da124eb-4da3-45e2-bc61-1924be259032"}, "owner").Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_DATA_SAMPLE,
		AssetKey:  storedDataSample.Key,
	}
	es.On("RegisterEvent", e).Once().Return(nil)

	err := service.UpdateDataSamples(updatedDataSample, "owner")

	assert.NoError(t, err, "Update of single valid assets should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateMultipleExistingDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	es := new(MockEventService)
	provider.On("GetEventService").Return(es)
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

	updatedDataSample := &asset.UpdateDataSamplesParam{
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

	es.On("RegisterEvent", mock.AnythingOfType("*asset.Event")).Times(2).Return(nil)

	err := service.UpdateDataSamples(updatedDataSample, "owner")

	assert.NoError(t, err, "Update of single valid assets should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateSingleNewDataSample(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dm := new(MockDataManagerService)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	updatedDataSample := &asset.UpdateDataSamplesParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"}, "owner").Return(nil).Once()
	dbal.On("GetDataSample", updatedDataSample.GetKeys()[0]).Return(&asset.DataSample{}, errors.New("sql Error")).Once()

	err := service.UpdateDataSamples(updatedDataSample, "owner")

	assert.Error(t, err, "Update of single unknown asset should fail")

	dbal.AssertExpectations(t)
}

func TestQueryDataSamples(t *testing.T) {
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

	dbal.On("QueryDataSamples", pagination).Return([]*asset.DataSample{&ds1, &ds2}, "nextPage", nil).Once()

	r, token, err := service.QueryDataSamples(pagination)

	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, ds1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestCheckSameManager(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	service := NewDataSampleService(provider)

	ds1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        false,
	}

	ds2 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
		TestOnly:        false,
	}

	dbal.On("GetDataSample", ds1.GetKey()).Return(ds1, nil)
	dbal.On("GetDataSample", ds2.GetKey()).Return(ds2, nil)

	err := service.CheckSameManager("9eef1e88-951a-44fb-944a-c3dbd1d72d85", []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84"})
	assert.NoError(t, err, "samples share a common manager")

	err = service.CheckSameManager("4da124eb-4da3-45e2-bc61-1924be259032", []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84"})
	assert.Error(t, err, "samples do not share a common manager")
}

func TestIsTestOnly(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetDataSampleDBAL").Return(dbal)
	service := NewDataSampleService(provider)

	ds1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		TestOnly:        true,
	}

	ds2 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
		TestOnly:        true,
	}

	ds3 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a85",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
		TestOnly:        false,
	}

	dbal.On("GetDataSample", ds1.GetKey()).Return(ds1, nil)
	dbal.On("GetDataSample", ds2.GetKey()).Return(ds2, nil)
	dbal.On("GetDataSample", ds3.GetKey()).Return(ds3, nil)

	testOnly, err := service.IsTestOnly([]string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a85"})
	assert.NoError(t, err, "check on usage should not fail")
	assert.False(t, testOnly)

	testOnly, err = service.IsTestOnly([]string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84"})
	assert.NoError(t, err, "check on usage should not fail")
	assert.True(t, testOnly)
}
