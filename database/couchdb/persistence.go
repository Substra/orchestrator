// Copyright 2020 Owkin Inc.
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

// Package couchdb implements the assets persistence layer.
// It relies on a couchdb backend.
package couchdb

import (
	"context"
	"encoding/json"
	"fmt"

	_ "github.com/go-kivik/couchdb/v3" // CouchDB driver
	"github.com/go-kivik/kivik/v3"
	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/utils"
)

// Persistence implements persistence.Database.
// It relies on a CouchDB database to store and query the assets.
type Persistence struct {
	db *kivik.DB
}

// NewPersistence connects to the given couchdb server and creates the database if needed.
func NewPersistence(ctx context.Context, dsn string, db string) (*Persistence, error) {
	client, err := kivik.New("couch", dsn)
	if err != nil {
		return nil, err
	}

	pong, err := client.Ping(ctx)
	if !pong {
		return nil, err
	}

	ensureDB(ctx, client, db)

	database := client.DB(ctx, db)

	return &Persistence{database}, err
}

// ensureDB makes sure the database exists
func ensureDB(ctx context.Context, client *kivik.Client, name string) error {
	all, err := client.AllDBs(ctx)
	if err != nil {
		return err
	}

	if utils.StringInSlice(all, name) {
		// Database already exists
		return nil
	}

	log.WithField("database", name).Info("Creating new database")
	err = client.CreateDB(ctx, name)
	return err
}

// Close underlying client
func (p *Persistence) Close(ctx context.Context) {
	p.db.Close(ctx)
	p.db.Client().Close(ctx)
}

// storedAsset wraps an asset to add docType metadata
type storedAsset struct {
	DocType string          `json:"doc_type"`
	Asset   json.RawMessage `json:"asset"`
	Rev     string          `json:"_rev,omitempty"`
}

func newStoredAsset(resource string, data []byte) (*storedAsset, error) {
	storedAsset := &storedAsset{
		DocType: resource,
		Asset:   data,
	}

	return storedAsset, nil
}

// PutState stores data or update existing data
func (p *Persistence) PutState(resource string, key string, data []byte) error {
	fullKey := getFullKey(resource, key)

	storedAsset, err := newStoredAsset(resource, data)
	if err != nil {
		return err
	}

	if r := p.db.Get(context.TODO(), getFullKey(resource, key)); r.Err == nil {
		// Document exists, revision is necessary to perform update
		storedAsset.Rev = r.Rev
	}

	_, err = p.db.Put(context.TODO(), fullKey, storedAsset)

	return err
}

// GetState fetches identified data
func (p *Persistence) GetState(resource string, key string) ([]byte, error) {
	r := p.db.Get(context.TODO(), getFullKey(resource, key))
	var buf []byte

	err := r.ScanDoc(buf)
	return buf, err
}

// GetAll retrieves all data for a resource kind
func (p *Persistence) GetAll(resource string) (result [][]byte, err error) {
	queryString := fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, resource)
	rows, err := p.db.Find(context.TODO(), queryString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		storedAsset := &storedAsset{}
		err := rows.ScanDoc(storedAsset)
		if err != nil {
			return nil, err
		}

		result = append(result, storedAsset.Asset)
	}

	rows.Close()

	return
}

func getFullKey(resource string, key string) string {
	return resource + ":" + key
}
