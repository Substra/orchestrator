package dbal

import (
	"context"
	"errors"
	"strconv"

	// migration driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/Masterminds/squirrel"
	// Database driver
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
)

const PgSortAsc = "ASC"
const PgSortDesc = "DESC"

// TransactionDBAL is a persistence.DBAL augmented with transaction capabilities.
// Its purpose is to be rolled back in case of error or committed at the end of a successful request.
type TransactionDBAL interface {
	persistence.DBAL
	Commit() error
	Rollback() error
}

// DBAL is the Database Abstraction Layer around asset storage
type DBAL struct {
	ctx     context.Context
	tx      pgx.Tx
	channel string
}

// Commit the changes to the underlying storage backend
func (d *DBAL) Commit() error {
	return d.tx.Commit(d.ctx)
}

// Rollback the changes so that the storage is left untouched
func (d *DBAL) Rollback() error {
	return d.tx.Rollback(d.ctx)
}

func getStatementBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func (d *DBAL) query(builder squirrel.Sqlizer) (pgx.Rows, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return d.tx.Query(d.ctx, query, args...)
}

func (d *DBAL) queryRow(builder squirrel.Sqlizer) (pgx.Row, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return d.tx.QueryRow(d.ctx, query, args...), nil
}

func (d *DBAL) exec(builder squirrel.Sqlizer) error { //nolint:golint,unused
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	_, err = d.tx.Exec(d.ctx, query, args...)
	return err
}

// AddNode implements persistence.NodeDBAL
func (d *DBAL) AddNode(node *asset.Node) error {
	stmt := `insert into "nodes" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(d.ctx, stmt, node.GetId(), node, d.channel)
	return err
}

// NodeExists implements persistence.NodeDBAL
func (d *DBAL) NodeExists(key string) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(id) from "nodes" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetAllNodes implements persistence.NodeDBAL
func (d *DBAL) GetAllNodes() ([]*asset.Node, error) {
	rows, err := d.tx.Query(d.ctx, `select "asset" from "nodes" where channel=$1`, d.channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*asset.Node

	for rows.Next() {
		node := new(asset.Node)

		err = rows.Scan(node)
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
	row := d.tx.QueryRow(d.ctx, `select "asset" from "nodes" where id=$1 and channel=$2`, id, d.channel)

	node := new(asset.Node)
	err := row.Scan(node)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("node", id)
		}
		return nil, err
	}

	return node, nil
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
