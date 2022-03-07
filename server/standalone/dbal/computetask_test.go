package dbal

import (
	"context"
	"errors"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/pashagolub/pgxmock"
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
		"single filter": {&asset.TaskQueryFilter{Worker: "mynode"}, "worker = $1", []interface{}{"mynode"}},
		"two filter":    {&asset.TaskQueryFilter{Worker: "mynode", Status: asset.ComputeTaskStatus_STATUS_DONE}, "worker = $1 AND status = $2", []interface{}{"mynode", asset.ComputeTaskStatus_STATUS_DONE.String()}},
		"three filter":  {&asset.TaskQueryFilter{Worker: "mynode", Status: asset.ComputeTaskStatus_STATUS_DONE, Category: asset.ComputeTaskCategory_TASK_TRAIN}, "worker = $1 AND status = $2 AND category = $3", []interface{}{"mynode", asset.ComputeTaskStatus_STATUS_DONE.String(), asset.ComputeTaskCategory_TASK_TRAIN.String()}},
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
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	keys := []string{"uuid1", "uuid2"}

	mock.ExpectQuery(`SELECT asset FROM compute_tasks`).WithArgs(testChannel, "uuid1", "uuid2").WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, err := dbal.GetComputeTasks(keys)
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetNoTask(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	mock.ExpectQuery(`select asset from "compute_tasks"`).
		WithArgs("uuid", testChannel).
		WillReturnError(pgx.ErrNoRows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, err = dbal.GetComputeTask("uuid")
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrNotFound, orcError.Kind)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryComputeTasks(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT asset FROM compute_tasks`).WithArgs(testChannel, "testWorker", asset.ComputeTaskStatus_STATUS_DONE.String()).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

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

func TestAddComputeTask(t *testing.T) {
	newTask := &asset.ComputeTask{
		Key:            "testTask",
		Category:       asset.ComputeTaskCategory_TASK_TEST,
		ComputePlanKey: "cpKey",
		Status:         asset.ComputeTaskStatus_STATUS_WAITING,
		Worker:         "testOrg",
		ParentTaskKeys: []string{"parent1", "parent2"},
	}

	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	// Insert task
	mock.ExpectExec(`insert into "compute_tasks"`).WithArgs(newTask.Key, testChannel, newTask.Category, newTask.ComputePlanKey, newTask.Status, newTask.Worker, newTask).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Insert parents
	mock.ExpectExec(`insert into compute_task_parents`).WithArgs("parent1", newTask.Key, 1).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec(`insert into compute_task_parents`).WithArgs("parent2", newTask.Key, 2).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.addTask(newTask)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddComputeTasks(t *testing.T) {
	newTasks := []*asset.ComputeTask{
		{
			Key:            "8d9fc421-15a6-4c3d-9082-3337a5436e83",
			Category:       asset.ComputeTaskCategory_TASK_TRAIN,
			ComputePlanKey: "899e7403-7e23-4c95-bb3f-7eb9e6d86b04",
			Status:         asset.ComputeTaskStatus_STATUS_WAITING,
			Worker:         "testOrg",
			ParentTaskKeys: []string{"46830c5b-5a42-4cd8-8c29-6b66cc1ef348", "46830c5b-5a42-4cd8-8c29-6b66cc1ef349"},
		},
		{
			Key:            "99d44ec9-d642-4afa-bad0-00dda84a6b9d",
			Category:       asset.ComputeTaskCategory_TASK_TEST,
			ComputePlanKey: "899e7403-7e23-4c95-bb3f-7eb9e6d86b04",
			Status:         asset.ComputeTaskStatus_STATUS_WAITING,
			Worker:         "testOrg",
			ParentTaskKeys: []string{"8d9fc421-15a6-4c3d-9082-3337a5436e83"},
		},
	}

	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	// Insert task
	mock.ExpectCopyFrom(`"compute_tasks"`, []string{"id", "channel", "category", "compute_plan_id", "status", "worker", "asset"}).WillReturnResult(2)
	// Insert parents relationships
	mock.ExpectCopyFrom(`"compute_task_parents"`, []string{"parent_task_id", "child_task_id", "position"}).WillReturnResult(3)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.addTasks(newTasks)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
