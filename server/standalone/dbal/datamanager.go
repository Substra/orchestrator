package dbal

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

// AddDataManager implements persistence.DataManagerDBAL
func (d *DBAL) AddDataManager(datamanager *asset.DataManager) error {
	stmt := `insert into "datamanagers" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(context.Background(), stmt, datamanager.GetKey(), datamanager, d.channel)
	return err
}

// UpdateDataManager implements persistence.DataManagerDBAL
func (d *DBAL) UpdateDataManager(datamanager *asset.DataManager) error {
	stmt := `update "datamanagers" set asset=$3 where id=$1 and channel=$2`
	_, err := d.tx.Exec(context.Background(), stmt, datamanager.GetKey(), d.channel, datamanager)
	return err
}

// DataManagerExists implements persistence.DataManagerDBAL
func (d *DBAL) DataManagerExists(key string) (bool, error) {
	row := d.tx.QueryRow(context.Background(), `select count(id) from "datamanagers" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetDataManager implements persistence.DataManagerDBAL
func (d *DBAL) GetDataManager(key string) (*asset.DataManager, error) {
	row := d.tx.QueryRow(context.Background(), `select "asset" from "datamanagers" where id=$1 and channel=$2`, key, d.channel)

	datamanager := new(asset.DataManager)
	err := row.Scan(&datamanager)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("datamanager not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return datamanager, nil
}

// QueryDataManagers implements persistence.DataManagerDBAL
func (d *DBAL) QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `select "asset" from "datamanagers" where channel=$3 order by created_at asc limit $1 offset $2`
	rows, err = d.tx.Query(context.Background(), query, p.Size+1, offset, d.channel)
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
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		bookmark = strconv.Itoa(offset + count)
	}

	return datamanagers, bookmark, nil
}
