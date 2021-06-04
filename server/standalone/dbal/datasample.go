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

package dbal

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

// DataSampleExists implements persistence.DataSampleDBAL
func (d *DBAL) DataSampleExists(key string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "datasamples" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// AddDataSample implements persistence.DataSampleDBAL
func (d *DBAL) AddDataSample(dataSample *asset.DataSample) error {
	stmt := `insert into "datasamples" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, dataSample.GetKey(), dataSample, d.channel)
	return err
}

// UpdateDataSample implements persistence.DataSampleDBAL
func (d *DBAL) UpdateDataSample(dataSample *asset.DataSample) error {
	stmt := `update "datasamples" set asset=$3 where id=$1 and channel=$2`
	_, err := d.tx.Exec(stmt, dataSample.GetKey(), d.channel, dataSample)
	return err
}

// GetDataSample implements persistence.DataSample
func (d *DBAL) GetDataSample(key string) (*asset.DataSample, error) {
	row := d.tx.QueryRow(`select "asset" from "datasamples" where id=$1 and channel=$2`, key, d.channel)

	datasample := new(asset.DataSample)
	err := row.Scan(&datasample)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("datasample not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return datasample, nil
}

// QueryDataSamples implements persistence.DataSample
func (d *DBAL) QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	var rows *sql.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `select "asset" from "datasamples" where channel=$3 order by created_at asc limit $1 offset $2`
	rows, err = d.tx.Query(query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var datasamples []*asset.DataSample
	var count int

	for rows.Next() {
		datasample := new(asset.DataSample)

		err = rows.Scan(&datasample)
		if err != nil {
			return nil, "", err
		}

		datasamples = append(datasamples, datasample)
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

	return datasamples, bookmark, nil
}

// GetDataSamplesKeysByDataManager implements persistence.DataSample
func (d *DBAL) GetDataSamplesKeysByDataManager(dataManagerKey string, testOnly bool) ([]string, error) {

	var rows *sql.Rows
	var err error

	testOnlyFilter := `not`
	if testOnly {
		testOnlyFilter = ``
	}

	query := `select "id" from "datasamples" where channel=$1 and (asset->'dataManagerKeys') ? $2 and ` + testOnlyFilter + ` (asset ? 'testOnly' and (asset->'testOnly')::boolean) order by created_at asc`

	rows, err = d.tx.Query(query, d.channel, dataManagerKey)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var datasampleKeys []string

	for rows.Next() {
		var datasampleKey string

		err = rows.Scan(&datasampleKey)
		if err != nil {
			return nil, err
		}
		datasampleKeys = append(datasampleKeys, datasampleKey)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return datasampleKeys, nil
}
