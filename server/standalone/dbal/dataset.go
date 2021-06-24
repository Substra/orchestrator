package dbal

import (
	"github.com/owkin/orchestrator/lib/asset"
)

// GetDataset implements persistence.DatasetDBAL
func (d *DBAL) GetDataset(id string) (*asset.Dataset, error) {

	datamanager, err := d.GetDataManager(id)

	if err != nil {
		return nil, err
	}

	trainDataSampleKeys, err := d.GetDataSamplesKeysByDataManager(id, false)
	if err != nil {
		return nil, err
	}

	testDataSampleKeys, err := d.GetDataSamplesKeysByDataManager(id, true)
	if err != nil {
		return nil, err
	}

	dataset := &asset.Dataset{
		DataManager:         datamanager,
		TrainDataSampleKeys: trainDataSampleKeys,
		TestDataSampleKeys:  testDataSampleKeys,
	}

	return dataset, nil
}
