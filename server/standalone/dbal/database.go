package dbal

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/utils"
)

type PgPool interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close()
}

// Database is a thin wrapper around PgPool.
// It handles the orchestrator specifics, such as DBAL creation.
type Database struct {
	Pool PgPool
}

type SQLLogger struct {
	verbose bool
}

func (l *SQLLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {

	sqlErr, hasError := data["err"]

	if hasError {
		// SQL errors should be logged at error level
		level = pgx.LogLevelError
	}

	if !(l.verbose || level == pgx.LogLevelError) {
		return
	}

	log := logger.Get(ctx).
		WithField("msg", msg).
		WithField("level", level)
	// Other available fields (omitted to keep logs readable):
	// - data["args"]: the query arguments (truncated if too long)
	// - data["time"]: the query execution time
	// - data["rowCount"]: number of rows returned, for SELECT statements
	// - data["pid"]

	if query, ok := data["sql"]; ok {
		log = log.WithField("sql", query)
	}

	if hasError {
		log = log.WithField("err", sqlErr)
	}

	switch level {
	case pgx.LogLevelTrace, pgx.LogLevelDebug:
		log.Debug("SQL")
	case pgx.LogLevelInfo:
		log.Info("SQL")
	case pgx.LogLevelWarn:
		log.Warn("SQL")
	case pgx.LogLevelError:
		log.Error("SQL")
	}
}

// InitDatabase opens a database connexion from given url.
func InitDatabase(databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	verbose, _ := utils.GetenvBool("LOG_SQL_VERBOSE")
	config.ConnConfig.Logger = &SQLLogger{verbose: verbose}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &Database{pool}, nil
}

// Close the connexion
func (d *Database) Close() {
	d.Pool.Close()
}

// BeginTransaction returns a new transaction.
// When readOnly is false the transaction is configured with SERIALIZABLE isolation level to protect against potential
// inconsistencies with concurrent requests.
func (d *Database) BeginTransaction(ctx context.Context, readOnly bool) (pgx.Tx, error) {
	logger.Get(ctx).WithField("ReadOnly", readOnly).Debug("new DB transaction")
	txOpts := pgx.TxOptions{
		IsoLevel: pgx.Serializable, // This level of isolation is the guarantee to always return consistent data
	}
	if readOnly {
		txOpts.AccessMode = pgx.ReadOnly
		txOpts.IsoLevel = pgx.ReadCommitted
	}
	return d.Pool.BeginTx(ctx, txOpts)
}
