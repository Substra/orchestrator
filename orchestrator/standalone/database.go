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

	"github.com/golang-migrate/migrate/v4"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/owkin/orchestrator/orchestrator/standalone/migrations"
)

// TransactionalDBALProvider describe an object able to return a TransactionalDBAL
type TransactionalDBALProvider interface {
	GetTransactionalDBAL() (TransactionDBAL, error)
}

// Database is a thin wrapper around sql.DB.
// It handles the orchestrator specifics, such as migrations and DBAL creation.
type Database struct {
	pool *sql.DB
}

// InitDatabase opens a database connexion from given url.
// It make sure the database has a usable schema by running migrations if there are any.
func InitDatabase(databaseURL string) (*Database, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	err = executeMigrations(databaseURL)
	if err != nil {
		return nil, err
	}

	return &Database{pool: db}, nil
}

func executeMigrations(databaseURL string) error {
	s := bindata.Resource(migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return migrations.Asset(name)
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

// GetTransactionalDBAL returns a new transactional DBAL
func (d *Database) GetTransactionalDBAL() (TransactionDBAL, error) {
	tx, err := d.pool.Begin()
	if err != nil {
		return nil, err
	}

	return &DBAL{tx: tx}, nil
}
