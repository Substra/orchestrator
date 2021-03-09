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

package ledger

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/common"
)

// DB is the distributed ledger persistence layer implementing persistence.DBAL
// This backend does not allow to read the current writes, they will only be commited after a successful response.
type DB struct {
	ccStub shim.ChaincodeStubInterface
	logger log.Entry
}

// NewDB creates a ledger.DB instance based on given stub
func NewDB(stub shim.ChaincodeStubInterface) *DB {
	logger := log.WithFields(
		log.F("db_backend", "ledger"),
	)

	return &DB{
		ccStub: stub,
		logger: logger,
	}
}

// storedAsset wraps an asset to add docType metadata
type storedAsset struct {
	DocType string          `json:"doc_type"`
	Asset   json.RawMessage `json:"asset"`
}

// putState stores data in the ledger
func (db *DB) putState(resource string, key string, data []byte) error {
	k := getFullKey(resource, key)
	logger := db.logger.WithFields(
		log.F("resource", resource),
		log.F("key", key),
		log.F("fullkey", k),
		log.F("data", data),
	)
	logger.Debug("put state")

	storedAsset := &storedAsset{
		DocType: resource,
		Asset:   data,
	}

	b, err := json.Marshal(storedAsset)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal stored asset")
		return err
	}
	logger.WithField("serialized stored asset", b).Debug("Marshalling successful")

	return db.ccStub.PutState(k, b)
}

// getState retrieves data for a given resource
func (db *DB) getState(resource string, key string) ([]byte, error) {
	k := getFullKey(resource, key)
	logger := db.logger.WithFields(
		log.F("resource", resource),
		log.F("key", key),
		log.F("fullkey", k),
	)
	logger.Debug("get state")

	b, err := db.ccStub.GetState(k)
	if err != nil {
		return nil, err
	}

	var stored storedAsset
	err = json.Unmarshal(b, &stored)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal stored asset")
		return nil, err
	}

	return stored.Asset, nil
}

// hasKey returns true if a resource with the same key already exists
func (db *DB) hasKey(resource string, key string) (bool, error) {
	k := getFullKey(resource, key)
	buff, err := db.ccStub.GetState(k)
	return buff != nil, err
}

// getAll fetch all data for a given resource kind
func (db *DB) getAll(resource string) (result [][]byte, err error) {
	logger := db.logger.WithFields(
		log.F("resource", resource),
	)
	logger.Debug("get all")

	queryString := fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, resource)

	resultsIterator, err := db.ccStub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var storedAsset storedAsset
		err = json.Unmarshal(queryResult.Value, &storedAsset)
		if err != nil {
			return nil, err
		}
		result = append(result, storedAsset.Asset)
	}

	return result, nil
}

func (db *DB) createIndex(index string, attributes []string) error {
	db.logger.WithField("index", index).WithField("attributes", attributes).Debug("Create index")
	compositeKey, err := db.ccStub.CreateCompositeKey(index, attributes)
	if err != nil {
		return err
	}
	value := []byte{0x00}
	if err = db.ccStub.PutState(compositeKey, value); err != nil {
		return err
	}
	return nil
}

// getIndexKeysWithPagination returns keys matching composite key values from the chaincode db.
func (db *DB) getIndexKeysWithPagination(index string, attributes []string, pageSize uint32, bookmark common.PaginationToken) ([]string, common.PaginationToken, error) {
	keys := []string{}
	db.logger.WithFields(
		log.F("index", index),
		log.F("attributes", attributes),
		log.F("bookmark", bookmark),
		log.F("pageSize", pageSize),
	).Debug("Get index keys")

	bookmark = json2couch(bookmark)

	iterator, metadata, err := db.ccStub.GetStateByPartialCompositeKeyWithPagination(index, attributes, int32(pageSize), bookmark)
	if err != nil {
		return nil, "", err
	}
	defer iterator.Close()
	for iterator.HasNext() {
		compositeKey, err := iterator.Next()
		if err != nil {
			return nil, "", err
		}
		_, keyParts, err := db.ccStub.SplitCompositeKey(compositeKey.Key)
		if err != nil {
			return nil, "", err
		}
		keys = append(keys, keyParts[len(keyParts)-1])
	}

	var nextPageToken string
	if metadata != nil {
		nextPageToken = couch2json(metadata.Bookmark)
	}

	return keys, nextPageToken, nil
}

func getFullKey(resource string, key string) string {
	return resource + ":" + key
}

// couch2json sanitizes input from CouchDB format to JSON-friendly format
func couch2json(in string) (out string) {
	if in == "" {
		return
	}
	out = strings.Replace(in, "\x00", "/", -1)
	out = strings.Replace(out, "\\u0000", "#", -1)
	out = strings.Replace(out, "\U0010ffff", "END", -1)
	return
}

// json2couch sanitizes input from JSON-friendly format to CouchDB format
func json2couch(in string) (out string) {
	if in == "" {
		return
	}
	out = strings.Replace(in, "/", "\x00", -1)
	out = strings.Replace(out, "#", "\\u0000", -1)
	out = strings.Replace(out, "END", "\U0010ffff", -1)
	return
}

// AddNode stores a new Node
func (db *DB) AddNode(node *assets.Node) error {
	nodeBytes, err := json.Marshal(node)
	if err != nil {
		return err
	}
	return db.putState(assets.NodeKind, node.GetId(), nodeBytes)
}

// GetNodes returns all known Nodes
func (db *DB) GetNodes() ([]*assets.Node, error) {
	b, err := db.getAll(assets.NodeKind)
	if err != nil {
		return nil, err
	}

	var nodes []*assets.Node

	for _, nodeBytes := range b {
		n := assets.Node{}
		err = json.Unmarshal(nodeBytes, &n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	return nodes, nil
}

// NodeExists test if a node with given ID is already stored
func (db *DB) NodeExists(id string) (bool, error) {
	return db.hasKey(assets.NodeKind, id)
}

// AddObjective stores a new objective
func (db *DB) AddObjective(obj *assets.Objective) error {
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = db.putState(assets.ObjectiveKind, obj.GetKey(), objBytes)
	if err != nil {
		return err
	}

	if err = db.createIndex("objective~owner~key", []string{"objective", obj.Owner, obj.Key}); err != nil {
		return err
	}

	return nil
}

// GetObjective retrieves an objective by its ID
func (db *DB) GetObjective(id string) (*assets.Objective, error) {
	o := assets.Objective{}

	b, err := db.getState(assets.ObjectiveKind, id)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// GetObjectives retrieves all objectives
func (db *DB) GetObjectives(p *common.Pagination) ([]*assets.Objective, common.PaginationToken, error) {
	elementsKeys, bookmark, err := db.getIndexKeysWithPagination("objective~owner~key", []string{"objective"}, p.Size, p.Token)
	if err != nil {
		return nil, "", err
	}

	db.logger.WithField("keys", elementsKeys).Debug("GetObjectives")

	var objectives []*assets.Objective
	for _, key := range elementsKeys {
		objective, err := db.GetObjective(key)
		if err != nil {
			return objectives, bookmark, err
		}
		objectives = append(objectives, objective)
	}

	return objectives, bookmark, nil
}
