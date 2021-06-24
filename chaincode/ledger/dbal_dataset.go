package ledger

import (
	"github.com/owkin/orchestrator/lib/asset"
)

// GetDataset implements persistence.DatasetDBAL
func (db *DB) GetDataset(id string) (*asset.Dataset, error) {
	datamanager, err := db.GetDataManager(id)
	if err != nil {
		return nil, err
	}

	trainDataSampleKeys, err := db.GetDataSamplesKeysByDataManager(id, false)
	if err != nil {
		return nil, err
	}

	testDataSampleKeys, err := db.GetDataSamplesKeysByDataManager(id, true)
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
