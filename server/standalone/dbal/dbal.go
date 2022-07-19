package dbal

import (
	"context"
	"strconv"

	"github.com/Masterminds/squirrel"
	// Database driver
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

const PgSortAsc = "ASC"
const PgSortDesc = "DESC"

// Conn is the database connection used by the DBAL.
type Conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	WaitForNotification(ctx context.Context) (*pgconn.Notification, error)
}

// DBAL is the Database Abstraction Layer around asset storage
type DBAL struct {
	ctx     context.Context
	tx      pgx.Tx
	conn    Conn
	channel string
}

func New(ctx context.Context, tx pgx.Tx, conn Conn, channel string) *DBAL {
	return &DBAL{ctx: ctx, tx: tx, conn: conn, channel: channel}
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

func (d *DBAL) exec(builder squirrel.Sqlizer) error {
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
