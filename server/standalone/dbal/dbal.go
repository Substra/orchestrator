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
	"context"
	"errors"
	"fmt"
	"strconv"

	// migration driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/Masterminds/squirrel"

	// Database driver
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
)

// TransactionDBAL is a persistence.DBAL augmented with transaction capabilities.
// It's purpose is to be rollbacked in case of error or commited at the end of a successful request.
type TransactionDBAL interface {
	persistence.DBAL
	Commit() error
	Rollback() error
}

// DBAL is the Database Abstraction Layer around asset storage
type DBAL struct {
	tx      pgx.Tx
	channel string
}

// Commit the changes to the underlying storage backend
func (d *DBAL) Commit() error {
	return d.tx.Commit(context.Background())
}

// Rollback the changes so that the storage is left untouched
func (d *DBAL) Rollback() error {
	return d.tx.Rollback(context.Background())
}

// AddNode implements persistence.NodeDBAL
func (d *DBAL) AddNode(node *asset.Node) error {
	stmt := `insert into "nodes" ("id", "channel") values ($1, $2)`
	_, err := d.tx.Exec(context.Background(), stmt, node.GetId(), d.channel)
	return err
}

// NodeExists implements persistence.NodeDBAL
func (d *DBAL) NodeExists(key string) (bool, error) {
	row := d.tx.QueryRow(context.Background(), `select count(id) from "nodes" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetAllNodes implements persistence.NodeDBAL
func (d *DBAL) GetAllNodes() ([]*asset.Node, error) {
	rows, err := d.tx.Query(context.Background(), `select "id" from "nodes" where channel=$1`, d.channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*asset.Node

	for rows.Next() {
		node := new(asset.Node)

		err = rows.Scan(&node.Id)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

// GetNode implements persistence.NodeDBAL
func (d *DBAL) GetNode(id string) (*asset.Node, error) {
	row := d.tx.QueryRow(context.Background(), `select "id" from "nodes" where id=$1 and channel=$2`, id, d.channel)

	node := new(asset.Node)
	err := row.Scan(&node.Id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("node not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return node, nil
}

// AddAlgo implements persistence.AlgoDBAL
func (d *DBAL) AddAlgo(algo *asset.Algo) error {
	stmt := `insert into "algos" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(context.Background(), stmt, algo.GetKey(), algo, d.channel)
	return err
}

// GetAlgo implements persistence.AlgoDBAL
func (d *DBAL) GetAlgo(key string) (*asset.Algo, error) {
	row := d.tx.QueryRow(context.Background(), `select "asset" from "algos" where id=$1 and channel=$2`, key, d.channel)

	algo := new(asset.Algo)
	err := row.Scan(&algo)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("algo not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return algo, nil
}

// QueryAlgos implements persistence.AlgoDBAL
func (d *DBAL) QueryAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select("asset").
		From("algos").
		Where(squirrel.Eq{"channel": d.channel}).
		OrderByClause("created_at ASC").
		Offset(uint64(offset)).
		Limit(uint64(p.Size + 1))

	if c != asset.AlgoCategory_ALGO_UNKNOWN {
		builder = builder.Where(squirrel.Eq{"asset->>'category'": c.String()})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err = d.tx.Query(context.Background(), query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var algos []*asset.Algo
	var count int

	for rows.Next() {
		algo := new(asset.Algo)

		err = rows.Scan(&algo)
		if err != nil {
			return nil, "", err
		}

		algos = append(algos, algo)
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

	return algos, bookmark, nil
}

// AlgoExists implements persistence.ObjectiveDBAL
func (d *DBAL) AlgoExists(key string) (bool, error) {
	row := d.tx.QueryRow(context.Background(), `select count(id) from "algos" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

func getOffset(token string) (int, error) {
	if token == "" {
		token = "0"
	}

	offset, err := strconv.Atoi(token)
	if err != nil {
		return 0, err
	}

	return offset, nil
}
