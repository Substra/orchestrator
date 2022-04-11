package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionsValue(t *testing.T) {
	permissions := &Permissions{
		Process:  &Permission{Public: true, AuthorizedIds: []string{"1", "2"}},
		Download: &Permission{Public: false, AuthorizedIds: []string{"4", "5"}},
	}

	value, err := permissions.Value()
	assert.NoError(t, err, "permissions serialization should not fail")

	scanned := new(Permissions)
	err = scanned.Scan(value)
	assert.NoError(t, err, "permissions scan should not fail")

	assert.Equal(t, permissions, scanned)
}

func TestAlgoCategoryValue(t *testing.T) {
	cat := AlgoCategory_ALGO_SIMPLE
	category := &cat

	value, err := category.Value()
	assert.NoError(t, err, "algo category serialization should not fail")

	scanned := new(AlgoCategory)
	err = scanned.Scan(value)
	assert.NoError(t, err, "algo category scan should not fail")

	assert.Equal(t, category, scanned)
}

func TestDataSampleValue(t *testing.T) {
	datasample := &DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "testOwner",
		TestOnly:        false,
	}

	value, err := datasample.Value()
	assert.NoError(t, err, "datasample serialization should not fail")

	scanned := new(DataSample)
	err = scanned.Scan(value)
	assert.NoError(t, err, "datasample scan should not fail")

	assert.Equal(t, datasample, scanned)
}

func TestDataManagerValue(t *testing.T) {
	datamanager := &DataManager{
		Name:  "test",
		Owner: "testOwner",
	}

	value, err := datamanager.Value()
	assert.NoError(t, err, "datamanager serialization should not fail")

	scanned := new(DataManager)
	err = scanned.Scan(value)
	assert.NoError(t, err, "datamanager scan should not fail")

	assert.Equal(t, datamanager, scanned)
}

func TestModelValue(t *testing.T) {
	model := &Model{
		Key:            "08bcb3b9-015c-4b6a-a9b5-033b3b324a7c",
		Category:       ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey: "08bcb3b9-015c-4b6a-a9b5-033b3b324a7d",
		Address: &Addressable{
			Checksum:       "c15918f80d920769904e92bae59cf4b926b362201ad686c2834403a08a19de16",
			StorageAddress: "http://somewhere.online",
		},
	}

	value, err := model.Value()
	assert.NoError(t, err, "model serialization should not fail")

	scanned := new(Model)
	err = scanned.Scan(value)
	assert.NoError(t, err, "model scan should not fail")

	assert.Equal(t, model, scanned)
}

func TestComputePlanValue(t *testing.T) {
	computeplan := &ComputePlan{
		Key: "08bcb3b9-015c-4b6a-a9b5-033b3b324a7c",
		Metadata: map[string]string{
			"test": "true",
		},
	}

	value, err := computeplan.Value()
	assert.NoError(t, err, "compute plan serialization should not fail")

	scanned := new(ComputePlan)
	err = scanned.Scan(value)
	assert.NoError(t, err, "compute plan scan should not fail")

	assert.Equal(t, computeplan, scanned)
}

func TestPerformanceValue(t *testing.T) {
	perf := &Performance{
		ComputeTaskKey:   "08bcb3b9-015c-4b6a-a9b5-033b3b324a7c",
		PerformanceValue: 0.43368,
	}

	value, err := perf.Value()
	assert.NoError(t, err, "performance serialization should not fail")

	scanned := new(Performance)
	err = scanned.Scan(value)
	assert.NoError(t, err, "performance scan should not fail")

	assert.Equal(t, perf, scanned)
}
