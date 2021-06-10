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
	"context"
	"testing"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataSampleFail(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	defer mock.Close(context.Background())

	mock.ExpectBegin()

	uid := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`select "asset" from "datasamples" where id=`).WithArgs(uid, testChannel)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{
		tx:      tx,
		channel: testChannel,
	}

	_, err = dbal.GetDataSample(uid)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
