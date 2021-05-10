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
	"database/sql"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
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

func (d *DBAL) QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error) {
	var rows *sql.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select("asset").
		From("models").
		Where(squirrel.Eq{"channel": d.channel}).
		OrderByClause("created_at ASC").
		Offset(uint64(offset)).
		Limit(uint64(p.Size + 1))

	if c != asset.ModelCategory_MODEL_UNKNOWN {
		builder = builder.Where(squirrel.Eq{"asset->>'category'": c.String()})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err = d.tx.Query(query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var models []*asset.Model
	var count int

	for rows.Next() {
		model := new(asset.Model)

		err = rows.Scan(&model)
		if err != nil {
			return nil, "", err
		}

		models = append(models, model)
		count++

		if count == int(p.Size) {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return models, bookmark, nil
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

func (d *DBAL) AddModel(model *asset.Model) error {
	stmt := `insert into "models" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, model.GetKey(), model, d.channel)
	return err
}

func (d *DBAL) UpdateModel(model *asset.Model) error {
	stmt := `update "models" set asset = $2 where id = $1 and channel = $3`
	_, err := d.tx.Exec(stmt, model.GetKey(), model, d.channel)
	return err
}
