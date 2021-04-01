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
func (db *DB) DataSampleExists(id string) (bool, error) {
	return db.hasKey(asset.DataSampleKind, id)
}

// AddDataSample implements persistence.DataSampleDBAL
func (db *DB) AddDataSample(dataSample *asset.DataSample) error {
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

	if err = db.createIndex("dataSample~owner~key", []string{"dataSample", dataSample.Owner, dataSample.Key}); err != nil {
		return err
	}

	for _, dataManagerKey := range dataSample.GetDataManagerKeys() {
		// create composite keys to find all dataSample associated with a dataManager
		if err = db.createIndex("dataSample~dataManager~key", []string{"dataSample", dataManagerKey, dataSample.GetKey()}); err != nil {
			return err
		}

		// create composite keys to find all dataSample associated with a dataManager that are for test only or not
		if err = db.createIndex("dataSample~dataManager~testOnly~key", []string{"dataSample", dataManagerKey, strconv.FormatBool(dataSample.GetTestOnly()), dataSample.GetKey()}); err != nil {
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
		if err = db.createIndex("dataSample~dataManager~key", []string{"dataSample", dataManagerKey, dataSample.GetKey()}); err != nil {
			return err
		}

		if err = db.createIndex("dataSample~dataManager~testOnly~key", []string{"dataSample", dataManagerKey, strconv.FormatBool(dataSample.GetTestOnly()), dataSample.GetKey()}); err != nil {
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
func (db *DB) GetDataSample(id string) (*asset.DataSample, error) {
	o := asset.DataSample{}

	b, err := db.getState(asset.DataSampleKind, id)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// GetDataSamples implements persistence.DataSampleDBAL
func (db *DB) GetDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	elementsKeys, bookmark, err := db.getIndexKeysWithPagination("dataSample~owner~key", []string{"dataSample"}, p.Size, p.Token)
	if err != nil {
		return nil, "", err
	}

	db.logger.WithField("keys", elementsKeys).Debug("GetDataSamples")

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
