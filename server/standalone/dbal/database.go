package dbal

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/standalone/migration"
	"github.com/owkin/orchestrator/utils"
)

type PgPool interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close()
}

// TransactionalDBALProvider describe an object able to return a TransactionalDBAL
type TransactionalDBALProvider interface {
	GetTransactionalDBAL(ctx context.Context, channel string, readOnly bool) (TransactionDBAL, error)
}

// Database is a thin wrapper around sql.DB.
// It handles the orchestrator specifics, such as migrations and DBAL creation.
type Database struct {
	pool PgPool
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

type MigrationsLogger struct {
}

func (ml *MigrationsLogger) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	// remove final newline
	if msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	log.WithField("context", "migrations").Info(msg)
}

func (ml *MigrationsLogger) Verbose() bool {
	return utils.GetLogLevelFromEnv() == log.DebugLevel
}

// InitDatabase opens a database connexion from given url.
// It make sure the database has a usable schema by running migrations if there are any.
func InitDatabase(databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	verbose, _ := utils.GetenvBool("METRICS_ENABLED")
	config.ConnConfig.Logger = &SQLLogger{verbose: verbose}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	err = executeMigrations(databaseURL)
	if err != nil {
		return nil, err
	}

	return &Database{pool}, nil
}

func executeMigrations(databaseURL string) error {
	d, err := iofs.New(migration.EmbeddedMigrations, ".")
	if err != nil {
		return err
	}
	// Prevent running migrations twice
	url := databaseURL + "&search_path=public"
	m, err := migrate.NewWithSourceInstance("iofs", d, url)
	if err != nil {
		return err
	}

	m.Log = &MigrationsLogger{}
	err = m.Up()
	// Treat no change as normal behavior
	if err != nil && errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}

// Close the connexion
func (d *Database) Close() {
	d.pool.Close()
}

// GetTransactionalDBAL returns a new transactional DBAL.
// The transaction is configured with SERIALIZABLE isolation level to protect against potential
// inconsistencies with concurrent requests.
func (d *Database) GetTransactionalDBAL(ctx context.Context, channel string, readOnly bool) (TransactionDBAL, error) {
	logger.Get(ctx).WithField("ReadOnly", readOnly).WithField("channel", channel).Debug("new DB transaction")
	txOpts := pgx.TxOptions{
		IsoLevel: pgx.Serializable, // This level of isolation is the guarantee to always return consistent data
	}
	if readOnly {
		txOpts.AccessMode = pgx.ReadOnly
		txOpts.IsoLevel = pgx.ReadCommitted
	}
	tx, err := d.pool.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}

	return &DBAL{ctx: ctx, tx: tx, channel: channel}, nil
}
