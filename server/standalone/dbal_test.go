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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testChannel = "testchannel"

func TestGetOffset(t *testing.T) {
	emptyOffset, err := getOffset("")
	assert.NoError(t, err)
	assert.Equal(t, 0, emptyOffset, "empty token should default to zero")

	valueOffset, err := getOffset("12")
	assert.NoError(t, err)
	assert.Equal(t, 12, valueOffset, "valued token should be parserd as int")
}

func TestQueryObjectives(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`select "asset" from "objectives"`).WithArgs(13, 0, testChannel).WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{tx, testChannel}

	res, bookmark, err := dbal.QueryObjectives(common.NewPagination("", 12))
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "", bookmark, "last page should be reached")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPaginatedQueryObjectives(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`select "asset" from "objectives"`).WithArgs(2, 0, testChannel).WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{tx, testChannel}

	res, bookmark, err := dbal.QueryObjectives(common.NewPagination("", 1))
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "1", bookmark, "There should be another page")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryAlgos(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT asset FROM algos`).WithArgs(testChannel, asset.AlgoCategory_ALGO_COMPOSITE.String()).WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{tx, testChannel}

	res, bookmark, err := dbal.QueryAlgos(asset.AlgoCategory_ALGO_COMPOSITE, common.NewPagination("", 12))
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "", bookmark, "last page should be reached")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPaginatedQueryAlgos(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT asset FROM algos`).WithArgs(testChannel, asset.AlgoCategory_ALGO_COMPOSITE.String()).WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{tx, testChannel}

	res, bookmark, err := dbal.QueryAlgos(asset.AlgoCategory_ALGO_COMPOSITE, common.NewPagination("", 1))
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "1", bookmark, "There should be another page")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetAlgoFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"})

	uid := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`select "asset" from "algos" where id=`).WithArgs(uid, testChannel).WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{
		tx:      tx,
		channel: testChannel,
	}

	_, err = dbal.GetAlgo(uid)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
