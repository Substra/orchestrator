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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"
)

// DB is the distributed ledger persistence layer implementing persistence.DBAL
// This backend does not allow to read the current writes, they will only be commited after a successful response.
type DB struct {
	context          context.Context
	ccStub           shim.ChaincodeStubInterface
	logger           log.Entry
	transactionState map[string]([]byte)
	transactionMutex sync.RWMutex
}

// NewDB creates a ledger.DB instance based on given stub
func NewDB(ctx context.Context, stub shim.ChaincodeStubInterface) *DB {
	logger := log.WithFields(
		log.F("db_backend", "ledger"),
	)

	return &DB{
		context:          ctx,
		ccStub:           stub,
		logger:           logger,
		transactionState: make(map[string]([]byte)),
		transactionMutex: sync.RWMutex{},
	}
}

// storedAsset wraps an asset to add docType metadata
type storedAsset struct {
	DocType string          `json:"doc_type"`
	Asset   json.RawMessage `json:"asset"`
}

// getTransactionState returns a copy of an object that has been updated or created during the transaction
func (db *DB) getTransactionState(key string) ([]byte, bool) {
	db.transactionMutex.RLock()
	defer db.transactionMutex.RUnlock()

	state, ok := db.transactionState[key]
	if !ok {
		return nil, false
	}

	b := make([]byte, len(state))
	copy(b, state)

	return b, true
}

// putTransactionState stores an object during a transaction lifetime
func (db *DB) putTransactionState(key string, b []byte) {
	db.transactionMutex.Lock()
	defer db.transactionMutex.Unlock()

	db.transactionState[key] = b
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
	logger.WithField("numBytes", len(b)).Debug("Marshalling successful")

	err = db.ccStub.PutState(k, b)

	if err != nil {
		return err
	}

	// TransactionState is updated to ensure that even if the data is not committed,
	// a further call to get this struct will return the updated one and not the original one
	db.putTransactionState(k, b)

	return nil

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

	b, ok := db.getTransactionState(k)

	var err error
	if !ok {
		b, err = db.ccStub.GetState(k)
		if err != nil {
			return nil, err
		}

		if len(b) == 0 {
			return nil, fmt.Errorf("%s not found: %w key: %s", resource, errors.ErrNotFound, key)
		}

		db.putTransactionState(k, b)
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

	resultsIterator, err := db.getQueryResult(queryString)
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

// validateQueryContext returns an error unless it is called in the context of an "Evaluate" transaction.
// validateQueryContext should ALWAYS be called from DB functions which perform CouchDB rich queries,
// and it should only be called from such functions. CouchDB rich queries provide no stability between chaincode
// execution and commit, hence they should only be executed in the context of "Evaluate" transactions.
// For more info, see https://hyperledger-fabric.readthedocs.io/en/release-1.4/couchdb_as_state_database.html#state-database-options
func (db *DB) validateQueryContext() error {
	isEvalTx := db.context.Value(ctxIsEvaluateTransaction)

	if isEvalTx == nil {
		return fmt.Errorf("missing key ctxIsEvaluateTransaction in transaction context: %w", errors.ErrInternalError)
	}

	if isEvalTx != true {
		fnName, err := utils.GetCaller(1)
		if err != nil {
			return err
		}
		return fmt.Errorf("function \"%s\" must be called from an \"Evaluate\" transaction: %w", fnName, errors.ErrInternalError)
	}

	return nil
}

func (db *DB) getQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	if err := db.validateQueryContext(); err != nil {
		return nil, err
	}

	return db.ccStub.GetQueryResult(query)
}

func (db *DB) getQueryResultWithPagination(query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	if err := db.validateQueryContext(); err != nil {
		return nil, nil, err
	}

	return db.ccStub.GetQueryResultWithPagination(query, pageSize, bookmark)
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

func (db *DB) deleteIndex(index string, attributes []string) error {
	compositeKey, err := db.ccStub.CreateCompositeKey(index, attributes)
	if err != nil {
		return err
	}
	return db.ccStub.DelState(compositeKey)
}

func (db *DB) updateIndex(index string, oldAttributes []string, newAttribues []string) error {
	if err := db.deleteIndex(index, oldAttributes); err != nil {
		return err
	}
	return db.createIndex(index, newAttribues)
}

func (db *DB) getIndexKeys(index string, attributes []string) ([]string, error) {
	keys := make([]string, 0)
	iterator, err := db.ccStub.GetStateByPartialCompositeKey(index, attributes)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	for iterator.HasNext() {
		compositeKey, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		_, keyParts, err := db.ccStub.SplitCompositeKey(compositeKey.Key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, keyParts[len(keyParts)-1])
	}
	return keys, nil
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
func (db *DB) AddNode(node *asset.Node) error {
	nodeBytes, err := json.Marshal(node)
	if err != nil {
		return err
	}
	return db.putState(asset.NodeKind, node.GetId(), nodeBytes)
}

// GetNodes returns all known Nodes
func (db *DB) GetNodes() ([]*asset.Node, error) {
	b, err := db.getAll(asset.NodeKind)
	if err != nil {
		return nil, err
	}

	var nodes []*asset.Node

	for _, nodeBytes := range b {
		n := asset.Node{}
		err = json.Unmarshal(nodeBytes, &n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	return nodes, nil
}

// GetNode returns a node by its ID
func (db *DB) GetNode(id string) (*asset.Node, error) {
	n := asset.Node{}

	b, err := db.getState(asset.NodeKind, id)
	if err != nil {
		return &n, err
	}

	err = json.Unmarshal(b, &n)
	return &n, err
}

// NodeExists test if a node with given ID is already stored
func (db *DB) NodeExists(id string) (bool, error) {
	return db.hasKey(asset.NodeKind, id)
}

// AddObjective stores a new objective
func (db *DB) AddObjective(obj *asset.Objective) error {
	exists, err := db.hasKey(asset.ObjectiveKind, obj.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("objective already exists: %w", errors.ErrConflict)
	}

	objBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = db.putState(asset.ObjectiveKind, obj.GetKey(), objBytes)
	if err != nil {
		return err
	}

	if err = db.createIndex("objective~owner~key", []string{asset.ObjectiveKind, obj.Owner, obj.Key}); err != nil {
		return err
	}

	return nil
}

// ObjectiveExists implements persistence.ObjectiveDBAL
func (db *DB) ObjectiveExists(id string) (bool, error) {
	return db.hasKey(asset.ObjectiveKind, id)
}

// GetObjective retrieves an objective by its ID
func (db *DB) GetObjective(id string) (*asset.Objective, error) {
	o := asset.Objective{}

	b, err := db.getState(asset.ObjectiveKind, id)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// GetObjectives retrieves all objectives
func (db *DB) GetObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	elementsKeys, bookmark, err := db.getIndexKeysWithPagination("objective~owner~key", []string{asset.ObjectiveKind}, p.Size, p.Token)
	if err != nil {
		return nil, "", err
	}

	db.logger.WithField("keys", elementsKeys).Debug("GetObjectives")

	var objectives []*asset.Objective
	for _, key := range elementsKeys {
		objective, err := db.GetObjective(key)
		if err != nil {
			return objectives, bookmark, err
		}
		objectives = append(objectives, objective)
	}

	return objectives, bookmark, nil
}

// AddAlgo stores a new algo
func (db *DB) AddAlgo(algo *asset.Algo) error {

	exists, err := db.hasKey(asset.AlgoKind, algo.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add algo: %w", errors.ErrConflict)
	}

	algoBytes, err := json.Marshal(algo)
	if err != nil {
		return err
	}
	err = db.putState(asset.AlgoKind, algo.GetKey(), algoBytes)
	if err != nil {
		return err
	}

	if err = db.createIndex("algo~owner~key", []string{asset.AlgoKind, algo.Owner, algo.Key}); err != nil {
		return err
	}

	if err = db.createIndex("algo~category~key", []string{"algo", algo.Category.String(), algo.Key}); err != nil {
		return err
	}

	return nil
}

// GetAlgo retrieves an algo by its ID
func (db *DB) GetAlgo(id string) (*asset.Algo, error) {
	a := asset.Algo{}

	b, err := db.getState(asset.AlgoKind, id)
	if err != nil {
		return &a, err
	}

	err = json.Unmarshal(b, &a)
	return &a, err
}

// GetAlgos retrieves all algos
func (db *DB) GetAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	logger := db.logger.WithFields(
		log.F("pagination", p),
		log.F("algo_category", c.String()),
	)
	logger.Debug("get algos")

	selector := couchTaskQuery{
		DocType: asset.AlgoKind,
	}

	assetFilter := map[string]interface{}{}
	if c != asset.AlgoCategory_ALGO_UNKNOWN {
		assetFilter["category"] = c.String()
	}
	if len(assetFilter) > 0 {
		selector.Asset = assetFilter
	}

	b, err := json.Marshal(selector)
	if err != nil {
		return nil, "", err
	}

	queryString := fmt.Sprintf(`{"selector":%s}}`, string(b))
	log.WithField("couchQuery", queryString).Debug("mango query")

	resultsIterator, bookmark, err := db.getQueryResultWithPagination(queryString, int32(p.Size), p.Token)
	if err != nil {
		return nil, "", err
	}
	defer resultsIterator.Close()

	algos := make([]*asset.Algo, 0)

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, "", err
		}
		var storedAsset storedAsset
		err = json.Unmarshal(queryResult.Value, &storedAsset)
		if err != nil {
			return nil, "", err
		}
		algo := &asset.Algo{}
		err = json.Unmarshal(storedAsset.Asset, algo)
		if err != nil {
			return nil, "", err
		}

		algos = append(algos, algo)
	}

	return algos, bookmark.Bookmark, nil
}

// AlgoExists implements persistence.ObjectiveDBAL
func (db *DB) AlgoExists(id string) (bool, error) {
	return db.hasKey(asset.AlgoKind, id)
}
