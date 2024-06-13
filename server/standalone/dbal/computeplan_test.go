package dbal

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
)

func TestGetComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"key", "owner", "creation_date", "cancelation_date", "failure_date", "tag", "name", "metadata"}).
		AddRow("uuid", "owner", time.Now(), nil, nil, "", "My compute plan", map[string]string{})

	mock.ExpectQuery(`SELECT key, owner, creation_date, cancelation_date, failure_date, tag, name, metadata`).
		WithArgs(testChannel, "uuid").
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, err = dbal.GetComputePlan("uuid")
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryComputePlans(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"key", "owner", "creation_date", "cancelation_date", "failure_date", "tag", "name", "metadata"}).
		AddRow("uuid", "owner", time.Now(), nil, nil, "", "My compute plan", map[string]string{})

	mock.ExpectQuery(`SELECT key,.* FROM compute_plans .* ORDER BY creation_date ASC, key ASC`).
		WithArgs(testChannel, "owner").
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	plans, _, err := dbal.QueryComputePlans(
		common.NewPagination("", 10),
		&asset.PlanQueryFilter{Owner: "owner"},
	)
	assert.NoError(t, err)
	assert.Len(t, plans, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryComputePlansNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"key", "owner", "creation_date", "cancelation_date", "failure_date", "tag", "name", "metadata"})

	mock.ExpectQuery(`SELECT key,.* FROM compute_plans .* ORDER BY creation_date ASC, key`).
		WithArgs(testChannel).
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, _, err = dbal.QueryComputePlans(
		common.NewPagination("", 10),
		nil,
	)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCancelComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	cpKey := "abc"
	cancelationDate, err := time.Parse("2006-01-02T15:04:05.000Z", "2021-02-03T04:05:06.007Z")
	assert.NoError(t, err)

	mock.ExpectBegin()

	mock.
		ExpectExec(`UPDATE compute_plans SET cancelation_date = $1 WHERE channel = $2 AND key = $3`).
		WithArgs(cancelationDate, testChannel, cpKey).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.CancelComputePlan(&asset.ComputePlan{Key: cpKey}, cancelationDate)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	cpKey := "abc"
	name := "My compute plan"

	mock.ExpectBegin()

	mock.
		ExpectExec(`UPDATE compute_plans SET name = $1 WHERE channel = $2 AND key = $3`).
		WithArgs(name, testChannel, cpKey).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	plan := &asset.ComputePlan{Key: cpKey, Name: name}
	err = dbal.SetComputePlanName(plan, plan.Name)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsPlanRunning(t *testing.T) {
	mock, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)

	cpKey := "abc"

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"status", "count"}).
		AddRow(asset.ComputeTaskStatus_STATUS_CANCELED.String(), 0).
		AddRow(asset.ComputeTaskStatus_STATUS_FAILED.String(), 0).
		AddRow(asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT.String(), 0).
		AddRow(asset.ComputeTaskStatus_STATUS_BUILDING.String(), 0).
		AddRow(asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS.String(), 0).
		AddRow(asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT.String(), 0).
		AddRow(asset.ComputeTaskStatus_STATUS_EXECUTING.String(), 1)
	mock.
		ExpectQuery(`SELECT status, COUNT(status) FROM compute_tasks WHERE channel = $1 AND compute_plan_key = $2 AND status IN ($3,$4,$5,$6,$7,$8,$9) GROUP BY status`).
		WithArgs(testChannel, cpKey, asset.ComputeTaskStatus_STATUS_CANCELED.String(), asset.ComputeTaskStatus_STATUS_FAILED.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT.String(), asset.ComputeTaskStatus_STATUS_BUILDING.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT.String(), asset.ComputeTaskStatus_STATUS_EXECUTING.String()).
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	isRunning, err := dbal.ArePlanTasksRunning(cpKey)
	assert.NoError(t, err)
	assert.True(t, isRunning)

	assert.NoError(t, mock.ExpectationsWereMet())
}
