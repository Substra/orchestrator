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

package dbal

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataSampleFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"})

	uid := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`select "asset" from "datasamples" where id=`).WithArgs(uid, testChannel).WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{
		tx:      tx,
		channel: testChannel,
	}

	_, err = dbal.GetDataSample(uid)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
