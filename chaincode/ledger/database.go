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

package ledger

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

var logger log.Entry

func init() {
	logger = log.WithFields(
		log.F("db_backend", "ledger"),
	)
}

// GetLedgerFromContext will return the ledger DB from invocation context
func GetLedgerFromContext(ctx contractapi.TransactionContextInterface) (persistence.Database, error) {
	stub := ctx.GetStub()

	return &DB{ccStub: stub}, nil
}

// DB is the distributed ledger persistence layer implementing persistence.Database
type DB struct {
	ccStub shim.ChaincodeStubInterface
}

// storedAsset wraps an asset to add docType metadata
type storedAsset struct {
	DocType string          `json:"doc_type"`
	Asset   json.RawMessage `json:"asset"`
}

// PutState stores data in the ledger
func (l *DB) PutState(resource string, key string, data []byte) error {
	k := getFullKey(resource, key)
	logger := logger.WithFields(
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

	return l.ccStub.PutState(k, b)
}

// GetState retrieves data for a given resource
func (l *DB) GetState(resource string, key string) ([]byte, error) {
	k := getFullKey(resource, key)
	logger := logger.WithFields(
		log.F("resource", resource),
		log.F("key", key),
		log.F("fullkey", k),
	)
	logger.Debug("get state")

	b, err := l.ccStub.GetState(k)
	if err != nil {
		return nil, err
	}

	var buf []byte
	err = json.Unmarshal(b, &buf)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal stored asset")
		return nil, err
	}

	return buf, nil
}

// GetAll fetch all data for a given resource kind
func (l *DB) GetAll(resource string) (result [][]byte, err error) {
	logger := logger.WithFields(
		log.F("resource", resource),
	)
	logger.Debug("get all")

	queryString := fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, resource)

	resultsIterator, err := l.ccStub.GetQueryResult(queryString)
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

func getFullKey(resource string, key string) string {
	return resource + ":" + key
}
