package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common/logger"
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
	logger := logger.Get(ctx)

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

// dbal indexes
const computePlanTaskStatusIndex = "computePlan~computePlanKey~status~task"
const computeTaskParentIndex = "computeTask~parentTask~key"
const modelTaskKeyIndex = "model~taskKey~modelKey"
const allNodesIndex = "nodes~id"

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
	logger := db.logger.WithField("query", query)
	if err := db.validateQueryContext(); err != nil {
		logger.WithError(err).Error("invalid context")
		return nil, err
	}

	return db.ccStub.GetQueryResult(query)
}

func (db *DB) getQueryResultWithPagination(query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	logger := db.logger.WithField("query", query)
	if err := db.validateQueryContext(); err != nil {
		logger.WithError(err).Error("invalid context")
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
	return db.ccStub.PutState(compositeKey, value)
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

func getFullKey(resource string, key string) string {
	return resource + ":" + key
}

// AddNode stores a new Node
func (db *DB) AddNode(node *asset.Node) error {
	nodeBytes, err := json.Marshal(node)
	if err != nil {
		return err
	}
	err = db.putState(asset.NodeKind, node.GetId(), nodeBytes)
	if err != nil {
		return err
	}

	return db.createIndex(allNodesIndex, []string{asset.NodeKind, node.Id})

}

// GetAllNodes returns all known Nodes
func (db *DB) GetAllNodes() ([]*asset.Node, error) {
	elementKeys, err := db.getIndexKeys(allNodesIndex, []string{asset.NodeKind})
	if err != nil {
		return nil, err
	}

	db.logger.WithField("numChildren", len(elementKeys)).Debug("GetAllNodes")

	nodes := []*asset.Node{}
	for _, id := range elementKeys {
		node, err := db.GetNode(id)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
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

	return nil
}

// GetAlgo retrieves an algo by its key
func (db *DB) GetAlgo(key string) (*asset.Algo, error) {
	a := asset.Algo{}

	b, err := db.getState(asset.AlgoKind, key)
	if err != nil {
		return &a, err
	}

	err = json.Unmarshal(b, &a)
	return &a, err
}

// QueryAlgos retrieves all algos
func (db *DB) QueryAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	logger := db.logger.WithFields(
		log.F("pagination", p),
		log.F("algo_category", c.String()),
	)
	logger.Debug("get algos")

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.AlgoKind,
		},
	}

	assetFilter := map[string]interface{}{}
	if c != asset.AlgoCategory_ALGO_UNKNOWN {
		assetFilter["category"] = c.String()
	}
	if len(assetFilter) > 0 {
		query.Selector.Asset = assetFilter
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, "", err
	}

	queryString := string(b)

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
func (db *DB) AlgoExists(key string) (bool, error) {
	return db.hasKey(asset.AlgoKind, key)
}

type couchAssetQuery struct {
	DocType string                 `json:"doc_type"`
	Asset   map[string]interface{} `json:"asset,omitempty"`
}

type richQuerySelector struct {
	Selector couchAssetQuery `json:"selector"`
	Fields   []string        `json:"fields,omitempty"`
}
