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

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

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
	docType         string
	serializedAsset []byte
}

// PutState stores data in the ledger
func (l *DB) PutState(resource string, key string, data []byte) error {
	k := getFullKey(resource, key)
	storedAsset := &storedAsset{
		docType:         resource,
		serializedAsset: data,
	}

	b, err := json.Marshal(storedAsset)
	if err != nil {
		return err
	}

	return l.ccStub.PutState(k, b)
}

// GetState retrieves data for a given resource
func (l *DB) GetState(resource string, key string) ([]byte, error) {
	k := getFullKey(resource, key)
	b, err := l.ccStub.GetState(k)
	if err != nil {
		return nil, err
	}

	var buf []byte
	err = json.Unmarshal(b, &buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// GetAll fetch all data for a given resource kind
func (l *DB) GetAll(resource string) ([][]byte, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"%s"}}`, resource)

	return l.getQueryResultForQueryString(queryString)
}

func getFullKey(resource string, key string) string {
	return resource + ":" + key
}

func (l *DB) getQueryResultForQueryString(queryString string) ([][]byte, error) {
	resultsIterator, err := l.ccStub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([][]byte, error) {
	var assets [][]byte
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
		assets = append(assets, storedAsset.serializedAsset)
	}

	return assets, nil
}
