package dbal

import (
	"context"
	"errors"
	"os"

	"github.com/golang-migrate/migrate/v4"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/standalone/migration"
)

// TransactionalDBALProvider describe an object able to return a TransactionalDBAL
type TransactionalDBALProvider interface {
	GetTransactionalDBAL(ctx context.Context, channel string, readOnly bool) (TransactionDBAL, error)
}

// Database is a thin wrapper around sql.DB.
// It handles the orchestrator specifics, such as migrations and DBAL creation.
type Database struct {
	pool *pgxpool.Pool
}

type SQLLogger struct {
	debug bool
}

func (l *SQLLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	if !l.debug && level <= pgx.LogLevelDebug {
		return
	}
	log := logger.Get(ctx).
		WithField("msg", msg).
		WithField("level", level)
		// Other available fields (omitted to keep logs readable):
		// - data["args"]: the query arguments (truncated if too long)
		// - data["time"]: the query execution time
		// - data["err"]: the SQL error text, if any
		// - data["rowCount"]: number of rows returned, for SELECT statements
		// - data["pid"]

	if query, ok := data["sql"]; ok {
		log = log.WithField("sql", query)
	}

	log.Debug("SQL")
}

// InitDatabase opens a database connexion from given url.
// It make sure the database has a usable schema by running migrations if there are any.
func InitDatabase(databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	_, logSQL := os.LookupEnv("LOG_SQL")
	config.ConnConfig.Logger = &SQLLogger{debug: logSQL}

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
	s := bindata.Resource(migration.AssetNames(),
		func(name string) ([]byte, error) {
			return migration.Asset(name)
		})

	d, err := bindata.WithInstance(s)
	if err != nil {
		return err
	}
	// Prevent running migrations twice
	url := databaseURL + "&search_path=public"
	m, err := migrate.NewWithSourceInstance("go-bindata", d, url)
	if err != nil {
		return err
	}

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
	}
	tx, err := d.pool.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, err
	}

	return &DBAL{ctx: ctx, tx: tx, channel: channel}, nil
}
