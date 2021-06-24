package ledger

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"
)

// DataSampleExists implements persistence.DataSampleDBAL
func (db *DB) DataSampleExists(key string) (bool, error) {
	return db.hasKey(asset.DataSampleKind, key)
}

func (db *DB) AddDataSamples(samples ...*asset.DataSample) error {
	for _, sample := range samples {
		err := db.addDataSample(sample)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) addDataSample(dataSample *asset.DataSample) error {
	exists, err := db.hasKey(asset.DataSampleKind, dataSample.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add datasample: %w", errors.ErrConflict)
	}

	dataSampleBytes, err := json.Marshal(dataSample)
	if err != nil {
		return err
	}

	err = db.putState(asset.DataSampleKind, dataSample.GetKey(), dataSampleBytes)
	if err != nil {
		return err
	}

	if err = db.createIndex("dataSample~owner~key", []string{asset.DataSampleKind, dataSample.Owner, dataSample.Key}); err != nil {
		return err
	}

	for _, dataManagerKey := range dataSample.GetDataManagerKeys() {
		// create composite keys to find all dataSample associated with a dataManager
		if err = db.createIndex("dataSample~dataManager~key", []string{asset.DataSampleKind, dataManagerKey, dataSample.GetKey()}); err != nil {
			return err
		}

		// create composite keys to find all dataSample associated with a dataManager that are for test only or not
		if err = db.createIndex("dataSample~dataManager~testOnly~key", []string{asset.DataSampleKind, dataManagerKey, strconv.FormatBool(dataSample.GetTestOnly()), dataSample.GetKey()}); err != nil {
			return err
		}
	}

	return nil
}

// UpdateDataSample implements persistence.DataSampleDBAL
func (db *DB) UpdateDataSample(dataSample *asset.DataSample) error {
	dataSampleBytes, err := json.Marshal(dataSample)
	if err != nil {
		return err
	}

	var currentDataSample *asset.DataSample
	currentDataSample, err = db.GetDataSample(dataSample.GetKey())
	if err != nil {
		// TODO define a better error than the sql error
		return err
	}

	newDataManagers := utils.Filter(dataSample.GetDataManagerKeys(), currentDataSample.GetDataManagerKeys())

	// We add indexes for the potential new DataManagerKeys
	for _, dataManagerKey := range newDataManagers {
		if err = db.createIndex("dataSample~dataManager~key", []string{asset.DataSampleKind, dataManagerKey, dataSample.GetKey()}); err != nil {
			return err
		}

		if err = db.createIndex("dataSample~dataManager~testOnly~key", []string{asset.DataSampleKind, dataManagerKey, strconv.FormatBool(dataSample.GetTestOnly()), dataSample.GetKey()}); err != nil {
			return err
		}
	}

	err = db.putState(asset.DataSampleKind, dataSample.GetKey(), dataSampleBytes)
	if err != nil {
		return err
	}

	return nil
}

// GetDataSample implements persistence.DataSampleDBAL
func (db *DB) GetDataSample(key string) (*asset.DataSample, error) {
	o := asset.DataSample{}

	b, err := db.getState(asset.DataSampleKind, key)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// QueryDataSamples implements persistence.DataSampleDBAL
func (db *DB) QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	elementsKeys, bookmark, err := db.getIndexKeysWithPagination("dataSample~owner~key", []string{asset.DataSampleKind}, p.Size, p.Token)
	if err != nil {
		return nil, "", err
	}

	db.logger.WithField("keys", elementsKeys).Debug("QueryDataSamples")

	var datasamples []*asset.DataSample
	for _, key := range elementsKeys {
		datasample, err := db.GetDataSample(key)
		if err != nil {
			return datasamples, bookmark, err
		}
		datasamples = append(datasamples, datasample)
	}

	return datasamples, bookmark, nil
}

// GetDataSamplesKeysByDataManager implements persistence.DataSampleDBAL
func (db *DB) GetDataSamplesKeysByDataManager(dataManagerKey string, testOnly bool) ([]string, error) {
	dataSampleKeys, err := db.getIndexKeys("dataSample~dataManager~testOnly~key", []string{asset.DataSampleKind, dataManagerKey, strconv.FormatBool(testOnly)})
	if err != nil {
		return nil, err
	}

	db.logger.WithField("keys", dataSampleKeys).Debug("GetDataSamplesKeysByDataManager")

	return dataSampleKeys, nil
}
