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
	"errors"
	"fmt"
	"strconv"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orchestrationErrors "github.com/owkin/orchestrator/lib/errors"
)

// AddDataManager implements persistence.DataManagerDBAL
func (d *DBAL) AddDataManager(datamanager *asset.DataManager) error {
	stmt := `insert into "datamanagers" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, datamanager.GetKey(), datamanager, d.channel)
	return err
}

// UpdateDataManager implements persistence.DataManagerDBAL
func (d *DBAL) UpdateDataManager(datamanager *asset.DataManager) error {
	stmt := `update "datamanagers" set asset=$3 where id=$1 and channel=$2`
	_, err := d.tx.Exec(stmt, datamanager.GetKey(), d.channel, datamanager)
	return err
}

// DataManagerExists implements persistence.DataManagerDBAL
func (d *DBAL) DataManagerExists(id string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "datamanagers" where id=$1 and channel=$2`, id, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetDataManager implements persistence.DataManagerDBAL
func (d *DBAL) GetDataManager(id string) (*asset.DataManager, error) {
	row := d.tx.QueryRow(`select "asset" from "datamanagers" where id=$1 and channel=$2`, id, d.channel)

	datamanager := new(asset.DataManager)
	err := row.Scan(&datamanager)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("datamanager not found: %w", orchestrationErrors.ErrNotFound)
		}
		return nil, err
	}

	return datamanager, nil
}

// GetDataManagers implements persistence.DataManagerDBAL
func (d *DBAL) GetDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	var rows *sql.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `select "asset" from "datamanagers" where channel=$3 order by created_at asc limit $1 offset $2`
	rows, err = d.tx.Query(query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var datamanagers []*asset.DataManager
	var count int

	for rows.Next() {
		datamanager := new(asset.DataManager)

		err = rows.Scan(&datamanager)
		if err != nil {
			return nil, "", err
		}

		datamanagers = append(datamanagers, datamanager)
		count++

		if count == int(p.Size) {
			break
		}
	}

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		bookmark = strconv.Itoa(offset + count)
	}

	return datamanagers, bookmark, nil
}
