package dbal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestToComputeTask(t *testing.T) {
	taskData := &asset.ComputeTask_Test{
		Test: &asset.TestTaskData{
			DataManagerKey: "dmkey",
			DataSampleKeys: []string{"dskey1", "dskey2"},
		},
	}

	marshalledTaskData, err := protojson.Marshal(&asset.ComputeTask{
		Data: taskData,
	})

	if err != nil {
		t.Fatalf("an error '%s' was not expected when marshalling task data", err)
	}

	algo := sqlAlgo{
		Key:         "algo_key",
		Name:        "algo_name",
		Description: asset.Addressable{},
		Algorithm:   asset.Addressable{},
		Permissions: asset.Permissions{
			Download: &asset.Permission{},
			Process:  &asset.Permission{},
		},
		Owner:        "algo_owner",
		CreationDate: time.Unix(111, 12).UTC(),
		Metadata:     map[string]string{},
	}
	ct := sqlComputeTask{
		Key:            "task_key",
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        algo.Key,
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
	assert.Equal(t, ct.AlgoKey, res.AlgoKey)
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

func makeTaskRows(taskKeys ...string) *pgxmock.Rows {
	permissions := []byte(`{"process": {"public": true}, "download": {"public": true}}`)
	res := pgxmock.NewRows([]string{"key", "compute_plan_key", "status", "category", "worker", "owner", "rank", "creation_date",
		"logs_permission", "task_data", "metadata", "algo_key", "parent_task_keys"})

	for _, key := range taskKeys {
		res = res.AddRow(key, "cp_key", "STATUS_WAITING", "TASK_TRAIN", "worker", "owner", int32(0), time.Unix(0, 100),
			[]byte("{}"), []byte("{}"), map[string]string{}, "algo_key", "algo_name", "https://description.foo",
			"d3ef77a", "https://algo.foo", "f3ed5a9", permissions, "owner", time.Unix(0, 100), map[string]string{}, []string{})
	}

	return res
}

func makeTaskInputRows(taskKeys ...string) *pgxmock.Rows {
	datasampleKeys := []string{"7b4d86a9-ab65-4d28-9358-8eb2edc952d9", "afc598a8-c01f-44bb-a082-e732e6aa875b"}
	res := pgxmock.NewRows([]string{"compute_task_key", "identifier", "asset_key", "parent_task_key", "parent_task_output_identifier"})

	for _, key := range taskKeys {
		for _, datasampleKey := range datasampleKeys {
			res.AddRow(key, "datasamples", datasampleKey, nil, nil)
		}
	}

	return res
}

func makeTaskOutputRows(taskKeys ...string) *pgxmock.Rows {
	res := pgxmock.NewRows([]string{"compute_task_key", "identifier", "permissions", "transient"})

	for _, key := range taskKeys {
		res.AddRow(key, "model", []byte("{}"), true)
	}

	return res
}

func TestTaskFilterToQuery(t *testing.T) {
	cases := map[string]struct {
		filter        *asset.TaskQueryFilter
		queryContains string
		params        []interface{}
	}{
		"empty":         {&asset.TaskQueryFilter{}, "", nil},
		"single filter": {&asset.TaskQueryFilter{Worker: "myorganization"}, "worker = $1", []interface{}{"myorganization"}},
		"two filter":    {&asset.TaskQueryFilter{Worker: "myorganization", Status: asset.ComputeTaskStatus_STATUS_DONE}, "worker = $1 AND status = $2", []interface{}{"myorganization", asset.ComputeTaskStatus_STATUS_DONE.String()}},
		"three filter":  {&asset.TaskQueryFilter{Worker: "myorganization", Status: asset.ComputeTaskStatus_STATUS_DONE, Category: asset.ComputeTaskCategory_TASK_TRAIN}, "worker = $1 AND status = $2 AND category = $3", []interface{}{"myorganization", asset.ComputeTaskStatus_STATUS_DONE.String(), asset.ComputeTaskCategory_TASK_TRAIN.String()}},
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
	assert.NoError(t, err)

	mock.ExpectBegin()

	keys := []string{"93733214-02b6-4d69-90a8-4e3518a63470", "32f7be0b-b432-41e5-8225-8c53580ccc58"}

	mock.ExpectQuery(`SELECT .* FROM expanded_compute_tasks`).
		WithArgs(testChannel, keys[0], keys[1]).
		WillReturnRows(makeTaskRows(keys[0], keys[1]))

	mock.ExpectQuery(`SELECT .* FROM compute_task_inputs`).
		WithArgs(keys[0], keys[1]).
		WillReturnRows(makeTaskInputRows(keys...))

	mock.ExpectQuery(`SELECT .* FROM compute_task_outputs`).
		WithArgs(keys[0], keys[1]).
		WillReturnRows(makeTaskOutputRows(keys...))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, err := dbal.GetComputeTasks(keys)
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNoTask(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

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

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryComputeTasks(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	keys := []string{"93733214-02b6-4d69-90a8-4e3518a63470", "32f7be0b-b432-41e5-8225-8c53580ccc58"}

	mock.ExpectQuery(`SELECT .* FROM expanded_compute_tasks`).
		WithArgs(testChannel, "testWorker", asset.ComputeTaskStatus_STATUS_DONE.String()).
		WillReturnRows(makeTaskRows(keys[0], keys[1]))

	mock.ExpectQuery(`SELECT .* FROM compute_task_inputs`).
		WithArgs(keys[0]).
		WillReturnRows(makeTaskInputRows(keys[0]))

	mock.ExpectQuery(`SELECT .* FROM compute_task_outputs`).
		WithArgs(keys[0]).
		WillReturnRows(makeTaskOutputRows(keys[0]))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, _, err := dbal.QueryComputeTasks(
		common.NewPagination("", 1),
		&asset.TaskQueryFilter{Worker: "testWorker", Status: asset.ComputeTaskStatus_STATUS_DONE},
	)
	assert.NoError(t, err)
	assert.Len(t, res, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComputeTask(t *testing.T) {
	newTask := &asset.ComputeTask{
		Key:            "8d9fc421-15a6-4c3d-9082-3337a5436e83",
		Category:       asset.ComputeTaskCategory_TASK_TEST,
		ComputePlanKey: "b16dcd88-32ca-4971-89a7-734b4ad1d778",
		Status:         asset.ComputeTaskStatus_STATUS_WAITING,
		Worker:         "testOrg",
		ParentTaskKeys: []string{"f7743332-17f5-4d20-9e29-55312a081c9d", "b09c3fbf-9f92-460b-a87c-37f7f3bd4c63"},
		Inputs: []*asset.ComputeTaskInput{
			{
				Identifier: "opener",
				Ref: &asset.ComputeTaskInput_AssetKey{
					AssetKey: "d57fcb21-9728-41c1-bbec-fcce919757e6",
				},
			},
			{
				Identifier: "datasamples",
				Ref: &asset.ComputeTaskInput_AssetKey{
					AssetKey: "5b9baa9c-89bb-48ba-b46e-311a2b426606",
				},
			},
		},
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {
				Permissions: &asset.Permissions{},
			},
		},
	}

	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	// Insert task
	mock.ExpectCopyFrom(`"compute_tasks"`,
		[]string{"key", "channel", "category", "algo_key", "owner", "compute_plan_key", "rank", "status", "worker", "creation_date", "logs_permission", "task_data", "metadata"}).
		WillReturnResult(1)
	// Insert parents relationships
	mock.ExpectCopyFrom(`"compute_task_parents"`, []string{"parent_task_key", "child_task_key", "position"}).WillReturnResult(2)
	// Insert task inputs
	mock.ExpectCopyFrom(`"compute_task_inputs"`, []string{"compute_task_key", "identifier", "position", "asset_key", "parent_task_key", "parent_task_output_identifier"}).WillReturnResult(2)
	// Insert task outputs
	mock.ExpectCopyFrom(`"compute_task_outputs"`, []string{"compute_task_key", "identifier", "permissions", "transient"}).WillReturnResult(1)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.AddComputeTasks(newTask)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
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
			Inputs: []*asset.ComputeTaskInput{
				{
					Identifier: "opener",
					Ref: &asset.ComputeTaskInput_AssetKey{
						AssetKey: "d57fcb21-9728-41c1-bbec-fcce919757e6",
					},
				},
				{
					Identifier: "datasamples",
					Ref: &asset.ComputeTaskInput_AssetKey{
						AssetKey: "5b9baa9c-89bb-48ba-b46e-311a2b426606",
					},
				},
			},
			Outputs: map[string]*asset.ComputeTaskOutput{
				"model": {
					Permissions: &asset.Permissions{},
				},
			},
		},
		{
			Key:            "99d44ec9-d642-4afa-bad0-00dda84a6b9d",
			Category:       asset.ComputeTaskCategory_TASK_TEST,
			ComputePlanKey: "899e7403-7e23-4c95-bb3f-7eb9e6d86b04",
			Status:         asset.ComputeTaskStatus_STATUS_WAITING,
			Worker:         "testOrg",
			ParentTaskKeys: []string{"8d9fc421-15a6-4c3d-9082-3337a5436e83"},
			Inputs: []*asset.ComputeTaskInput{
				{
					Identifier: "model",
					Ref: &asset.ComputeTaskInput_AssetKey{
						AssetKey: "f16a376b-e896-45f3-bea6-e3388c766335",
					},
				},
			},
			Outputs: map[string]*asset.ComputeTaskOutput{
				"local": {
					Permissions: &asset.Permissions{},
				},
				"shared": {
					Permissions: &asset.Permissions{},
				},
			},
		},
	}

	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	// Insert task
	mock.ExpectCopyFrom(`"compute_tasks"`,
		[]string{"key", "channel", "category", "algo_key", "owner", "compute_plan_key", "rank", "status", "worker", "creation_date", "logs_permission", "task_data", "metadata"}).
		WillReturnResult(2)
	// Insert parents relationships
	mock.ExpectCopyFrom(`"compute_task_parents"`, []string{"parent_task_key", "child_task_key", "position"}).WillReturnResult(3)
	// Insert task inputs
	mock.ExpectCopyFrom(`"compute_task_inputs"`, []string{"compute_task_key", "identifier", "position", "asset_key", "parent_task_key", "parent_task_output_identifier"}).WillReturnResult(3)
	// Insert task outputs
	mock.ExpectCopyFrom(`"compute_task_outputs"`, []string{"compute_task_key", "identifier", "permissions", "transient"}).WillReturnResult(3)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.AddComputeTasks(newTasks...)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryComputeTasksNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	keys := []string{"93733214-02b6-4d69-90a8-4e3518a63470", "32f7be0b-b432-41e5-8225-8c53580ccc58"}

	mock.ExpectQuery(`SELECT .* FROM expanded_compute_tasks`).
		WithArgs(testChannel).
		WillReturnRows(makeTaskRows(keys[0], keys[1]))

	mock.ExpectQuery(`SELECT .* FROM compute_task_inputs`).
		WithArgs(keys[0]).
		WillReturnRows(makeTaskInputRows(keys[0]))

	mock.ExpectQuery(`SELECT .* FROM compute_task_outputs`).
		WithArgs(keys[0]).
		WillReturnRows(makeTaskOutputRows(keys[0]))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, _, err = dbal.QueryComputeTasks(
		common.NewPagination("", 1),
		nil,
	)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComputeTaskOutputAsset(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	output := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              "taskKey",
		ComputeTaskOutputIdentifier: "identifierOut",
		AssetKind:                   asset.AssetKind_ASSET_ALGO,
		AssetKey:                    "assetKey",
	}

	mock.ExpectExec(`INSERT INTO compute_task_output_assets \(.*\) SELECT .* FROM compute_task_output_assets WHERE compute_task_key = \$\d AND compute_task_output_identifier = \$\d`).
		WithArgs(output.ComputeTaskKey, output.ComputeTaskOutputIdentifier, output.AssetKind, output.AssetKey, output.ComputeTaskKey, output.ComputeTaskOutputIdentifier).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.AddComputeTaskOutputAsset(output)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetComputeTaskOutputAssets(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	res := pgxmock.NewRows([]string{"asset_kind", "asset_key"}).
		AddRow("ASSET_MODEL", "6d313ee9-3ea6-4ceb-abaa-eac9643863a6")

	mock.ExpectQuery(`SELECT asset_kind, asset_key FROM compute_task_output_assets WHERE compute_task_key = \$\d AND compute_task_output_identifier = \$\d ORDER BY position ASC`).
		WithArgs("e9133395-f3c1-4407-96cc-e1681815bea3", "model").
		WillReturnRows(res)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	expectedOutput := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              "e9133395-f3c1-4407-96cc-e1681815bea3",
		ComputeTaskOutputIdentifier: "model",
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    "6d313ee9-3ea6-4ceb-abaa-eac9643863a6",
	}

	outputs, err := dbal.GetComputeTaskOutputAssets("e9133395-f3c1-4407-96cc-e1681815bea3", "model")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	assert.Len(t, outputs, 1)
	assert.Equal(t, expectedOutput, outputs[0])
}
