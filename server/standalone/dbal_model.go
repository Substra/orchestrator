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

package standalone

import (
	"github.com/owkin/orchestrator/lib/asset"
)

func (d *DBAL) GetModel(key string) (*asset.Model, error) {
	row := d.tx.QueryRow(`select asset from "models" where id=$1 and channel=$2`, key, d.channel)

	model := new(asset.Model)
	err := row.Scan(model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (d *DBAL) ModelExists(key string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "models" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

func (d *DBAL) GetComputeTaskOutputModels(key string) ([]*asset.Model, error) {
	rows, err := d.tx.Query(`select asset from "models" where asset->>'computeTaskKey' = $1 and channel=$2`, key, d.channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := []*asset.Model{}
	for rows.Next() {
		model := new(asset.Model)
		err := rows.Scan(model)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

func (d *DBAL) AddModel(model *asset.Model) error {
	stmt := `insert into "models" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, model.GetKey(), model, d.channel)
	return err
}

func (d *DBAL) UpdateModel(model *asset.Model) error {
	stmt := `update "models" set asset = $2 where id = $1 and channel = $3)`
	_, err := d.tx.Exec(stmt, model.GetKey(), model, d.channel)
	return err
}
