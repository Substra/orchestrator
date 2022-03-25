package dbal

import (
	"context"
	"strconv"

	"github.com/Masterminds/squirrel"
	// migration driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Database driver
	"github.com/jackc/pgx/v4"
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
