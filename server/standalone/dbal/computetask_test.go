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
	"github.com/Masterminds/squirrel"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskFilterToQuery(t *testing.T) {
	cases := map[string]struct {
		filter        *asset.TaskQueryFilter
		queryContains string
		params        []interface{}
	}{
		"empty":         {&asset.TaskQueryFilter{}, "", nil},
		"single filter": {&asset.TaskQueryFilter{Worker: "mynode"}, "asset->>'worker' = $1", []interface{}{"mynode"}},
		"two filter":    {&asset.TaskQueryFilter{Worker: "mynode", Status: asset.ComputeTaskStatus_STATUS_DONE}, "asset->>'worker' = $1 AND asset->>'status' = $2", []interface{}{"mynode", asset.ComputeTaskStatus_STATUS_DONE.String()}},
		"three filter":  {&asset.TaskQueryFilter{Worker: "mynode", Status: asset.ComputeTaskStatus_STATUS_DONE, Category: asset.ComputeTaskCategory_TASK_TRAIN}, "asset->>'worker' = $1 AND asset->>'status' = $2 AND asset->>'category' = $3", []interface{}{"mynode", asset.ComputeTaskStatus_STATUS_DONE.String(), asset.ComputeTaskCategory_TASK_TRAIN.String()}},
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := pgDialect.Select("asset").From("compute_tasks")
			builder = taskFilterToQuery(c.filter, builder)
			query, params, err := builder.ToSql()
			assert.NoError(t, err)
			assert.Contains(t, query, c.queryContains)
			assert.Equal(t, c.params, params)
		})
	}
}

func TestGetTasks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	keys := []string{"uuid1", "uuid2"}

	mock.ExpectQuery(`SELECT asset FROM compute_tasks`).WithArgs(testChannel, "uuid1", "uuid2").WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{tx, testChannel}

	res, err := dbal.GetComputeTasks(keys)
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryComputeTasks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT asset FROM compute_tasks`).WithArgs(testChannel, "testWorker", asset.ComputeTaskStatus_STATUS_DONE.String()).WillReturnRows(rows)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbal := &DBAL{tx, testChannel}

	res, _, err := dbal.QueryComputeTasks(
		common.NewPagination("", 1),
		&asset.TaskQueryFilter{Worker: "testWorker", Status: asset.ComputeTaskStatus_STATUS_DONE},
	)
	assert.NoError(t, err)
	assert.Len(t, res, 1)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}