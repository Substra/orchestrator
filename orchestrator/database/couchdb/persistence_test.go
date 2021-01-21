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

package couchdb

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/go-kivik/kivik/v3/driver"
	"github.com/go-kivik/kivikmock/v3"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type subAsset struct {
	Kind string `json:"kind"`
	Bool bool   `json:"bool"`
}

type asset struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Sub  []subAsset `json:"sub"`
}

func TestAssetMarshal(t *testing.T) {
	testAsset := asset{
		ID:   "test",
		Name: "test value",
		Sub:  []subAsset{{Kind: "sub1", Bool: true}},
	}

	serializedAsset, err := json.Marshal(testAsset)
	require.Nil(t, err, "marshalling test asset should not fail")

	storedAsset, err := newStoredAsset("test", serializedAsset)
	require.Nil(t, err, "creating storedAsset should not fail")

	serializedWrapper, err := json.Marshal(storedAsset)
	require.Nil(t, err, "marshalling should not fail")

	expected := `{"doc_type":"test","asset":{"id":"test","name":"test value","sub":[{"kind":"sub1","bool":true}]}}`

	assert.Equal(t, expected, string(serializedWrapper), "Serialized wrapper should not reserialize the underlying asset")
}

func TestAssetUnmarshal(t *testing.T) {
	serialized := `{"doc_type":"test","asset":{"id":"test","name":"test value","sub":[{"kind":"sub1","bool":true}]}}`

	stored := &storedAsset{}

	err := json.Unmarshal([]byte(serialized), stored)
	require.Nil(t, err, "unmarshalling stored asset should not fail")

	a := &asset{}
	err = json.Unmarshal(stored.Asset, a)

	assert.Nil(t, err, "unmarshalling the underlying asset should not fail")
	assert.Equal(t, "sub1", a.Sub[0].Kind, "Asset should be successfully unmarshalled")
}

func TestGetFullKey(t *testing.T) {
	k := getFullKey("resource", "id")

	assert.Equal(t, "resource:id", k, "key should be prefixed with resource type")
}

func TestEnsureDBWithExistingDB(t *testing.T) {
	client, mock, err := kivikmock.New()
	require.NoError(t, err)

	mock.ExpectDBExists().WithName("test").WillReturn(true)

	err = ensureDB(context.TODO(), client, "test")
	assert.NoError(t, err)
}

func TestEnsureDBCreatesDB(t *testing.T) {
	client, mock, err := kivikmock.New()
	require.NoError(t, err)

	mock.ExpectDBExists().WithName("test").WillReturn(false)
	mock.ExpectCreateDB().WillReturnError(nil)

	err = ensureDB(context.TODO(), client, "test")
	assert.NoError(t, err)
}

func TestPutStateNewEntry(t *testing.T) {
	client, mock, err := kivikmock.New()
	require.NoError(t, err)

	newDoc := storedAsset{
		DocType: "test",
		Asset:   []byte{},
	}

	db := mock.NewDB()

	mock.ExpectDB().WillReturn(db)

	db.ExpectGet().WithDocID(getFullKey("test", "uuid")).WillReturnError(errors.New("not found"))
	db.ExpectPut().WithDocID(getFullKey("test", "uuid")).WithDoc(newDoc)

	persistence := &Persistence{client.DB(context.TODO(), "testdb")}

	err = persistence.PutState("test", "uuid", []byte{})
	assert.NoError(t, err, "PutState should not fail")
}

func TestPutStateExistingEntry(t *testing.T) {
	client, mock, err := kivikmock.New()
	require.NoError(t, err)

	existingDoc := `{"_id": "test:uuid", "_rev": "1-d16971943ea33664ac6fe1241ea6388e", "asset": {}}`
	updatedDoc := storedAsset{
		DocType: "test",
		Asset:   []byte("{}"),
		Rev:     "1-d16971943ea33664ac6fe1241ea6388e",
	}

	db := mock.NewDB()

	mock.ExpectDB().WillReturn(db)

	db.ExpectGet().WithDocID(getFullKey("test", "uuid")).WillReturn(kivikmock.DocumentT(t, existingDoc))
	db.ExpectPut().WithDocID(getFullKey("test", "uuid")).WithDoc(updatedDoc).WillReturn("2-whatever")

	persistence := &Persistence{client.DB(context.TODO(), "testdb")}

	err = persistence.PutState("test", "uuid", []byte("{}"))
	assert.NoError(t, err, "PutState should not fail")
}

func TestGetState(t *testing.T) {

	client, mock, err := kivikmock.New()
	require.NoError(t, err)
	db := mock.NewDB()

	mock.ExpectDB().WillReturn(db)

	doc := `{
		"_id": "nodes:test",
		"_rev": "1-d16971943ea33664ac6fe1241ea6388e",
		"doc_type": "nodes",
		"asset": {
			"id": "test"
		}
	}`

	db.ExpectGet().WithDocID(getFullKey("nodes", "test")).WillReturn(kivikmock.DocumentT(t, doc))

	persistence := &Persistence{client.DB(context.TODO(), "testdb")}

	b, err := persistence.GetState("nodes", "test")
	assert.NoError(t, err)

	a := asset{}
	json.Unmarshal(b, &a)

	assert.Equal(t, "test", a.ID, "Asset should be retrieved from storage")
}

func TestGetAll(t *testing.T) {
	client, mock, err := kivikmock.New()
	require.NoError(t, err)
	db := mock.NewDB()

	mock.ExpectDB().WillReturn(db)

	doc1 := `{
		"_id": "nodes:test1",
		"_rev": "1-d16971943ea33664ac6fe1241ea6388e",
		"doc_type": "nodes",
		"asset": {
			"id": "test1"
		}
	}`
	doc2 := `{
		"_id": "nodes:test2",
		"_rev": "1-d16971943ea33664ac6fe1241ea6388e",
		"doc_type": "nodes",
		"asset": {
			"id": "test2"
		}
	}`
	rows := kivikmock.NewRows().
		AddRow(&driver.Row{ID: "nodes:test1", Doc: []byte(doc1)}).
		AddRow(&driver.Row{ID: "nodes:test2", Doc: []byte(doc2)})

	db.ExpectFind().WithQuery(`{"selector":{"doc_type":"nodes"}}`).WillReturn(rows)

	persistence := &Persistence{client.DB(context.TODO(), "testdb")}
	res, err := persistence.GetAll("nodes")

	assert.NoError(t, err)
	assert.Len(t, res, 2, "GetAll should return all entities")

	for i, b := range res {
		var a asset
		err = json.Unmarshal(b, &a)
		assert.NoError(t, err, "Unmarshalling asset should not fail")
		assert.Equal(t, "test"+strconv.Itoa(i+1), a.ID, "Asset should have proper ID")
	}
}

func TestNewPersistence(t *testing.T) {
	client, mock, err := kivikmock.New()
	require.NoError(t, err)
	db := mock.NewDB()

	mock.ExpectPing().WillReturn(true)
	mock.ExpectDBExists().WithName("test").WillReturn(true)
	mock.ExpectDB().WillReturn(db)

	p, err := newPersistence(context.TODO(), client, "test")

	assert.NoError(t, err, "newPersistence should not fail")
	assert.Implements(t, (*persistence.Database)(nil), p, "newPersistence should return a Database object")
}

func TestNewPersistenceNoPong(t *testing.T) {
	client, mock, err := kivikmock.New()
	require.NoError(t, err)

	mock.ExpectPing().WillReturnError(errors.New("no pong"))

	_, err = newPersistence(context.TODO(), client, "test")

	assert.Error(t, err, "newPersistence should fail")
}
