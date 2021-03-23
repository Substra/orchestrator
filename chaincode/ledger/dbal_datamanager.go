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

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
)

// AddDataManager stores a new DataManager
func (db *DB) AddDataManager(datamanager *asset.DataManager) error {
	exists, err := db.hasKey(asset.DataManagerKind, datamanager.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add datamanager: %w", errors.ErrConflict)
	}

	dataManagerBytes, err := json.Marshal(datamanager)
	if err != nil {
		return err
	}

	err = db.putState(asset.DataManagerKind, datamanager.GetKey(), dataManagerBytes)
	if err != nil {
		return err
	}

	// Create a composite key to DataManagers associated with an objective.
	// It does not make sense to create a composite key if there is no objective set,
	// which is why we do not set it if the objective key is empty.
	if datamanager.GetObjectiveKey() != "" {
		err = db.createIndex("dataManager~objective~key", []string{"dataManager", datamanager.GetObjectiveKey(), datamanager.GetKey()})
		if err != nil {
			return err
		}
	}

	// Create a composite key to find DataManagers associated with an owner
	err = db.createIndex("dataManager~owner~key", []string{"dataManager", datamanager.GetOwner(), datamanager.GetKey()})
	if err != nil {
		return err
	}

	return nil
}

// UpdateDataManager implements persistence.DataManagerDBAL
func (db *DB) UpdateDataManager(datamanager *asset.DataManager) error {
	dataManagerBytes, err := json.Marshal(datamanager)
	if err != nil {
		return err
	}

	err = db.putState(asset.DataManagerKind, datamanager.GetKey(), dataManagerBytes)
	if err != nil {
		return err
	}

	// Here it makes sense to create a composite key because the only thing we can update on the
	// DataManager is the objective key and you can only update it if there is no current objective key set.
	err = db.createIndex("dataManager~objective~key", []string{"dataManager", datamanager.GetObjectiveKey(), datamanager.GetKey()})
	if err != nil {
		return err
	}

	return nil
}

// DataManagersExists implements persistence.DataManagerDBAL
func (db *DB) DataManagersExists(id string) (bool, error) {
	return db.hasKey(asset.DataManagerKind, id)
}

// GetDataManager implements persistence.DataManagerDBAL
func (db *DB) GetDataManager(id string) (*asset.DataManager, error) {
	d := asset.DataManager{}

	b, err := db.getState(asset.DataManagerKind, id)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &d)
	return &d, err
}

// GetDataManagers implements persistence.DataManagerDBAL
func (db *DB) GetDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	elementsKeys, bookmark, err := db.getIndexKeysWithPagination("dataManager~owner~key", []string{"dataManager"}, p.Size, p.Token)
	if err != nil {
		return nil, "", err
	}

	db.logger.WithField("keys", elementsKeys).Debug("GetDataManagers")

	var datamanagers []*asset.DataManager
	for _, key := range elementsKeys {
		datamanager, err := db.GetDataManager(key)
		if err != nil {
			return datamanagers, bookmark, err
		}
		datamanagers = append(datamanagers, datamanager)
	}

	return datamanagers, bookmark, nil
}
