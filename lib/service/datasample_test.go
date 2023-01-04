package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterSingleDataSample(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	dm := new(MockDataManagerAPI)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	provider.On("GetTimeService").Return(ts)
	service := NewDataSampleService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	datasample := &asset.NewDataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	storedDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		CreationDate:    timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("DataSampleExists", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83").Return(false, nil).Once()
	dbal.On("AddDataSamples", storedDataSample).Return(nil).Once()
	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_DATA_SAMPLE,
		AssetKey:  storedDataSample.Key,
		Asset:     &asset.Event_DataSample{DataSample: storedDataSample},
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	_, err := service.RegisterDataSamples([]*asset.NewDataSample{datasample}, "owner")

	assert.NoError(t, err, "Registration of valid datasample should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterSingleDataSampleUnknownDataManager(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	dm := new(MockDataManagerAPI)
	provider := newMockedProvider()
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	datasample := &asset.NewDataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(errors.New("unknown datamanager")).Once()

	_, err := service.RegisterDataSamples([]*asset.NewDataSample{datasample}, "owner")

	assert.Error(t, err, "Registration of datasample with invalid datamanager key should fail")

	dbal.AssertExpectations(t)
}

func TestRegisterMultipleDataSamples(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	dm := new(MockDataManagerAPI)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	provider.On("GetTimeService").Return(ts)
	service := NewDataSampleService(provider)

	ts.On("GetTransactionTime").Twice().Return(time.Unix(1337, 0))

	datasamples := []*asset.NewDataSample{
		{
			Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
		{
			Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
			DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
			Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	storedDataSample1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		CreationDate:    timestamppb.New(time.Unix(1337, 0)),
	}

	storedDataSample2 := &asset.DataSample{
		Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
		Checksum:        "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		CreationDate:    timestamppb.New(time.Unix(1337, 0)),
	}

	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"}, "owner").Return(nil).Times(2)
	dbal.On("DataSampleExists", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83").Return(false, nil).Once()
	dbal.On("DataSampleExists", "0b4b4466-9a81-4084-9bab-80939b78addd").Return(false, nil).Once()
	dbal.On("AddDataSamples", storedDataSample1, storedDataSample2).Return(nil).Once()

	es.On("RegisterEvents", mock.AnythingOfType("*asset.Event"), mock.AnythingOfType("*asset.Event")).Once().Return(nil)

	_, err := service.RegisterDataSamples(datasamples, "owner")

	assert.NoError(t, err, "Registration of multiple valid assets should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestUpdateSingleExistingDataSample(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	dm := new(MockDataManagerAPI)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	existingDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
	}

	updatedDataSample := &asset.UpdateDataSamplesParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		DataManagerKeys: []string{"4da124eb-4da3-45e2-bc61-1924be259032"},
	}

	storedDataSample := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
	}

	dbal.On("GetDataSample", existingDataSample.GetKey()).Return(existingDataSample, nil).Once()
	dbal.On("UpdateDataSample", storedDataSample).Return(nil).Once()
	dm.On("CheckOwner", []string{"4da124eb-4da3-45e2-bc61-1924be259032"}, "owner").Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_DATA_SAMPLE,
		AssetKey:  storedDataSample.Key,
		Asset:     &asset.Event_DataSample{DataSample: storedDataSample},
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	err := service.UpdateDataSamples(updatedDataSample, "owner")

	assert.NoError(t, err, "Update of single valid assets should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateMultipleExistingDataSample(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	dm := new(MockDataManagerAPI)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataSampleDBAL").Return(dbal)
	provider.On("GetDataManagerService").Return(dm)
	service := NewDataSampleService(provider)

	existingDataSample1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
	}

	existingDataSample2 := &asset.DataSample{
		Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
	}

	updatedDataSample := &asset.UpdateDataSamplesParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "0b4b4466-9a81-4084-9bab-80939b78addd"},
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
	}

	storedDataSample1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
	}

	storedDataSample2 := &asset.DataSample{
		Key:             "0b4b4466-9a81-4084-9bab-80939b78addd",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
	}

	dbal.On("GetDataSample", existingDataSample1.GetKey()).Return(existingDataSample1, nil).Once()
	dbal.On("GetDataSample", existingDataSample2.GetKey()).Return(existingDataSample2, nil).Once()
	dm.On("CheckOwner", []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"}, "owner").Return(nil).Once()
	dbal.On("UpdateDataSample", storedDataSample1).Return(nil).Once()
	dbal.On("UpdateDataSample", storedDataSample2).Return(nil).Once()

	es.On("RegisterEvents", mock.AnythingOfType("*asset.Event")).Times(2).Return(nil)

	err := service.UpdateDataSamples(updatedDataSample, "owner")

	assert.NoError(t, err, "Update of single valid assets should not fail")

	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateSingleNewDataSample(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	dm := new(MockDataManagerAPI)
	provider := newMockedProvider()
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
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetDataSampleDBAL").Return(dbal)
	service := NewDataSampleService(provider)

	ds1 := asset.DataSample{
		Key:      "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
	}
	ds2 := asset.DataSample{
		Key:      "9eef1e88-951a-44fb-944a-c3dbd1d72d85",
	}

	pagination := common.NewPagination("", 10)

	filter := (*asset.DataSampleQueryFilter)(nil)

	dbal.On("QueryDataSamples", pagination, filter).Return([]*asset.DataSample{&ds1, &ds2}, "nextPage", nil).Once()

	r, token, err := service.QueryDataSamples(pagination, filter)

	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, ds1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestCheckSameManager(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetDataSampleDBAL").Return(dbal)
	service := NewDataSampleService(provider)

	ds1 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "owner",
	}

	ds2 := &asset.DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85", "4da124eb-4da3-45e2-bc61-1924be259032"},
		Owner:           "owner",
	}

	dbal.On("GetDataSample", ds1.GetKey()).Return(ds1, nil)
	dbal.On("GetDataSample", ds2.GetKey()).Return(ds2, nil)

	err := service.CheckSameManager("9eef1e88-951a-44fb-944a-c3dbd1d72d85", []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84"})
	assert.NoError(t, err, "samples share a common manager")

	err = service.CheckSameManager("4da124eb-4da3-45e2-bc61-1924be259032", []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84"})
	assert.Error(t, err, "samples do not share a common manager")
}

func TestGetDataSample(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetDataSampleDBAL").Return(dbal)
	service := NewDataSampleService(provider)

	ds1 := asset.DataSample{
		Key:      "4c67ad88-309a-48b4-8bc4-c2e2c1a87a84",
		Owner:    "owner",
	}

	dbal.On("GetDataSample", ds1.GetKey()).Return(&ds1, nil).Once()

	o, err := service.GetDataSample("4c67ad88-309a-48b4-8bc4-c2e2c1a87a84")
	require.Nil(t, err)
	assert.Equal(t, o.Owner, ds1.Owner)
	dbal.AssertExpectations(t)
}
