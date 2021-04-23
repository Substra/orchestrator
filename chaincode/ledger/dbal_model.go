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
	"github.com/owkin/orchestrator/lib/errors"
)

func (db *DB) GetModel(key string) (*asset.Model, error) {
	model := new(asset.Model)

	b, err := db.getState(asset.ModelKind, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (db *DB) ModelExists(key string) (bool, error) {
	exists, err := db.hasKey(asset.ModelKind, key)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (db *DB) GetTaskModels(key string) ([]*asset.Model, error) {
	elementKeys, err := db.getIndexKeys("model~taskKey~modelKey", []string{asset.ModelKind, key})
	if err != nil {
		return nil, err
	}

	db.logger.WithField("numChildren", len(elementKeys)).Debug("GetTaskModels")

	models := []*asset.Model{}
	for _, modelKey := range elementKeys {
		model, err := db.GetModel(modelKey)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

func (db *DB) AddModel(model *asset.Model) error {
	exists, err := db.hasKey(asset.ModelKind, model.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add model: %w", errors.ErrConflict)
	}
	bytes, err := json.Marshal(model)
	if err != nil {
		return err
	}

	err = db.putState(asset.ModelKind, model.GetKey(), bytes)
	if err != nil {
		return err
	}

	if err := db.createIndex("model~taskKey~modelKey", []string{asset.ModelKind, model.GetComputeTaskKey(), model.GetKey()}); err != nil {
		return err
	}

	return nil
}
