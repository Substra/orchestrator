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
	"testing"

	"github.com/go-kivik/kivikmock/v3"
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
