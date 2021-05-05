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

	// migration driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/Masterminds/squirrel"

	// Database driver
	_ "github.com/jackc/pgx/v4/stdlib"
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
	tx      *sql.Tx
	channel string
}

// Commit the changes to the underlying storage backend
func (d *DBAL) Commit() error {
	return d.tx.Commit()
}

// Rollback the changes so that the storage is left untouched
func (d *DBAL) Rollback() error {
	return d.tx.Rollback()
}

// AddNode implements persistence.NodeDBAL
func (d *DBAL) AddNode(node *asset.Node) error {
	stmt := `insert into "nodes" ("id", "channel") values ($1, $2)`
	_, err := d.tx.Exec(stmt, node.GetId(), d.channel)
	return err
}

// NodeExists implements persistence.NodeDBAL
func (d *DBAL) NodeExists(key string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "nodes" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetNodes implements persistence.NodeDBAL
func (d *DBAL) GetNodes() ([]*asset.Node, error) {
	rows, err := d.tx.Query(`select "id" from "nodes" where channel=$1`, d.channel)
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

	return nodes, nil
}

// GetNode implements persistence.NodeDBAL
func (d *DBAL) GetNode(id string) (*asset.Node, error) {
	row := d.tx.QueryRow(`select "asset" from "nodes" where id=$1 and channel=$2`, id, d.channel)

	node := new(asset.Node)
	err := row.Scan(&node)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("node not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return node, nil
}

// AddObjective implements persistence.ObjectiveDBAL
func (d *DBAL) AddObjective(obj *asset.Objective) error {
	stmt := `insert into "objectives" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, obj.GetKey(), obj, d.channel)

	return err
}

// GetObjective implements persistence.ObjectiveDBAL
func (d *DBAL) GetObjective(key string) (*asset.Objective, error) {
	row := d.tx.QueryRow(`select "asset" from "objectives" where id=$1 and channel=$2`, key, d.channel)

	objective := new(asset.Objective)
	err := row.Scan(&objective)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("objective not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return objective, nil
}

// ObjectiveExists implements persistence.ObjectiveDBAL
func (d *DBAL) ObjectiveExists(key string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "objectives" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetObjectives implements persistence.ObjectiveDBAL
func (d *DBAL) GetObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	var rows *sql.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `select "asset" from "objectives" where channel=$3 order by created_at asc limit $1 offset $2`
	rows, err = d.tx.Query(query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var objectives []*asset.Objective
	var count int

	for rows.Next() {
		objective := new(asset.Objective)

		err = rows.Scan(&objective)
		if err != nil {
			return nil, "", err
		}

		objectives = append(objectives, objective)
		count++

		if count == int(p.Size) {
			break
		}
	}

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return objectives, bookmark, nil
}

// AddAlgo implements persistence.AlgoDBAL
func (d *DBAL) AddAlgo(algo *asset.Algo) error {
	stmt := `insert into "algos" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, algo.GetKey(), algo, d.channel)
	return err
}

// GetAlgo implements persistence.AlgoDBAL
func (d *DBAL) GetAlgo(key string) (*asset.Algo, error) {
	row := d.tx.QueryRow(`select "asset" from "algos" where id=$1 and channel=$2`, key, d.channel)

	algo := new(asset.Algo)
	err := row.Scan(&algo)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("algo not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return algo, nil
}

// GetAlgos implements persistence.AlgoDBAL
func (d *DBAL) GetAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	var rows *sql.Rows
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

	rows, err = d.tx.Query(query, args...)
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

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return algos, bookmark, nil
}

// AlgoExists implements persistence.ObjectiveDBAL
func (d *DBAL) AlgoExists(key string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "algos" where id=$1 and channel=$2`, key, d.channel)

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
