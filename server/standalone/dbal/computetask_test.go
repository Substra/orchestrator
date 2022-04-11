package dbal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestToComputeTask(t *testing.T) {
	taskData := &asset.ComputeTask_Test{
		Test: &asset.TestTaskData{
			DataManagerKey: "dmkey",
			DataSampleKeys: []string{"dskey1", "dskey2"},
			MetricKeys:     []string{"mkey"},
		},
	}

	marshalledTaskData, err := protojson.Marshal(&asset.ComputeTask{
		Data: taskData,
	})

	if err != nil {
		t.Fatalf("an error '%s' was not expected when marshalling task data", err)
	}

	ct := sqlComputeTask{
		Key:      "task_key",
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
		Algo: sqlAlgo{
			Key:         "algo_key",
			Name:        "algo_name",
			Category:    asset.AlgoCategory_ALGO_COMPOSITE,
			Description: asset.Addressable{},
			Algorithm:   asset.Addressable{},
			Permissions: asset.Permissions{
				Download: &asset.Permission{},
				Process:  &asset.Permission{},
			},
			Owner:        "algo_owner",
			CreationDate: time.Unix(111, 12).UTC(),
			Metadata:     map[string]string{},
		},
		Owner:          "owner",
		ComputePlanKey: "cp_key",
		ParentTaskKeys: []string{},
		Rank:           0,
		Status:         asset.ComputeTaskStatus_STATUS_DOING,
		Worker:         "worker",
		CreationDate:   time.Unix(100, 10).UTC(),
		LogsPermission: asset.Permission{
			Public:        false,
			AuthorizedIds: []string{},
		},
		Data:     marshalledTaskData,
		Metadata: map[string]string{},
	}

	res, err := ct.toComputeTask()
	assert.NoError(t, err)
	assert.Equal(t, ct.Key, res.Key)
	assert.Equal(t, ct.Category, res.Category)
	assert.Equal(t, ct.Algo.Key, res.Algo.Key)
	assert.Equal(t, ct.Owner, res.Owner)
	assert.Equal(t, ct.ComputePlanKey, res.ComputePlanKey)
	assert.Equal(t, ct.ParentTaskKeys, res.ParentTaskKeys)
	assert.Equal(t, ct.Rank, res.Rank)
	assert.Equal(t, ct.Status, res.Status)
	assert.Equal(t, ct.Worker, res.Worker)
	assert.Equal(t, ct.CreationDate, res.CreationDate.AsTime())
	assert.Equal(t, &ct.LogsPermission, res.LogsPermission)
	assert.Equal(t, ct.Metadata, res.Metadata)
	assert.Equal(t, taskData, res.Data)
}

func makeTaskRows() *pgxmock.Rows {
	permissions := []byte(`{"process": {"public": true}, "download": {"public": true}}`)
	return pgxmock.NewRows([]string{"key", "compute_plan_key", "status", "category", "worker", "owner", "rank", "creation_date",
		"logs_permission", "task_data", "metadata", "algo_key", "algo_name", "algo_category", "algo_description_address",
		"algo_description_checksum", "algo_algorithm_address", "algo_algorithm_checksum", "algo_permissions", "algo_owner",
		"algo_creation_date", "algo_metadata", "parent_task_keys"}).
		AddRow("key1", "cp_key", "STATUS_WAITING", "TASK_TRAIN", "worker", "owner", int32(0), time.Unix(0, 100),
			[]byte("{}"), []byte("{}"), map[string]string{}, "algo_key", "algo_name", "ALGO_SIMPLE", "https://description.foo",
			"d3ef77a", "https://algo.foo", "f3ed5a9", permissions, "owner", time.Unix(0, 100), map[string]string{}, []string{}).
		AddRow("key2", "cp_key", "STATUS_WAITING", "TASK_TRAIN", "worker", "owner", int32(0), time.Unix(0, 100),
			[]byte("{}"), []byte("{}"), map[string]string{}, "algo_key", "algo_name", "ALGO_SIMPLE", "https://description.foo",
			"d3ef77a", "https://algo.foo", "f3ed5a9", permissions, "owner", time.Unix(0, 100), map[string]string{}, []string{})
}

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

	keys := []string{"uuid1", "uuid2"}
	mock.ExpectQuery(`SELECT .* FROM expanded_compute_tasks`).
		WithArgs(testChannel, keys[0], keys[1]).
		WillReturnRows(makeTaskRows())

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

	mock.ExpectQuery(`SELECT .* FROM expanded_compute_tasks`).
		WithArgs(testChannel, "uuid").
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

	mock.ExpectQuery(`SELECT .* FROM expanded_compute_tasks`).
		WithArgs(testChannel, "testWorker", asset.ComputeTaskStatus_STATUS_DONE.String()).
		WillReturnRows(makeTaskRows())

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
		Key:            "8d9fc421-15a6-4c3d-9082-3337a5436e83",
		Category:       asset.ComputeTaskCategory_TASK_TEST,
		ComputePlanKey: "b16dcd88-32ca-4971-89a7-734b4ad1d778",
		Status:         asset.ComputeTaskStatus_STATUS_WAITING,
		Worker:         "testOrg",
		ParentTaskKeys: []string{"f7743332-17f5-4d20-9e29-55312a081c9d", "b09c3fbf-9f92-460b-a87c-37f7f3bd4c63"},
	}

	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	// Insert task
	mock.ExpectCopyFrom(`"compute_tasks"`,
		[]string{"key", "channel", "category", "algo_key", "owner", "compute_plan_key", "rank", "status", "worker", "creation_date", "logs_permission", "task_data", "metadata"}).
		WillReturnResult(1)
	// Insert parents relationships
	mock.ExpectCopyFrom(`"compute_task_parents"`, []string{"parent_task_key", "child_task_key", "position"}).WillReturnResult(2)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.AddComputeTasks(newTask)
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
	mock.ExpectCopyFrom(`"compute_tasks"`,
		[]string{"key", "channel", "category", "algo_key", "owner", "compute_plan_key", "rank", "status", "worker", "creation_date", "logs_permission", "task_data", "metadata"}).
		WillReturnResult(2)
	// Insert parents relationships
	mock.ExpectCopyFrom(`"compute_task_parents"`, []string{"parent_task_key", "child_task_key", "position"}).WillReturnResult(3)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.AddComputeTasks(newTasks...)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryComputeTasksNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT .* FROM expanded_compute_tasks`).
		WithArgs(testChannel).
		WillReturnRows(makeTaskRows())

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, _, err = dbal.QueryComputeTasks(
		common.NewPagination("", 1),
		nil,
	)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
