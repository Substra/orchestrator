package dbal

import (
	"context"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"key", "owner", "delete_intermediary_models", "creation_date", "cancelation_date", "tag", "name", "metadata", "task_count", "waiting_count", "todo_count", "doing_count", "canceled_count", "failed_count", "done_count"}).
		AddRow("uuid", "owner", false, time.Now(), nil, "", "My compute plan", map[string]string{}, uint32(21), uint32(1), uint32(2), uint32(3), uint32(4), uint32(5), uint32(6))

	mock.ExpectQuery(`SELECT key, owner, delete_intermediary_models, creation_date, cancelation_date, tag, name, metadata, task_count, waiting_count, todo_count, doing_count, canceled_count, failed_count, done_count`).
		WithArgs(testChannel, "uuid").
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	plan, err := dbal.GetComputePlan("uuid")
	assert.NoError(t, err)

	assert.Equal(t, uint32(21), plan.TaskCount)
	assert.Equal(t, uint32(1), plan.WaitingCount)
	assert.Equal(t, uint32(2), plan.TodoCount)
	assert.Equal(t, uint32(3), plan.DoingCount)
	assert.Equal(t, uint32(4), plan.CanceledCount)
	assert.Equal(t, uint32(5), plan.FailedCount)
	assert.Equal(t, uint32(6), plan.DoneCount)
	assert.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_FAILED, plan.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRawComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"key", "owner", "delete_intermediary_models", "creation_date", "cancelation_date", "tag", "name", "metadata"}).
		AddRow("uuid", "owner", false, time.Now(), nil, "", "My compute plan", map[string]string{})

	mock.ExpectQuery(`SELECT key, owner, delete_intermediary_models, creation_date, cancelation_date, tag, name, metadata FROM compute_plans`).
		WithArgs(testChannel, "uuid").
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	plan, err := dbal.GetRawComputePlan("uuid")
	assert.NoError(t, err)

	assert.Equal(t, uint32(0), plan.TaskCount)
	assert.Equal(t, uint32(0), plan.DoneCount)
	assert.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN, plan.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryComputePlans(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"key", "owner", "delete_intermediary_models", "creation_date", "cancelation_date", "tag", "name", "metadata", "task_count", "waiting_count", "todo_count", "doing_count", "canceled_count", "failed_count", "done_count"}).
		AddRow("uuid", "owner", false, time.Now(), nil, "", "My compute plan", map[string]string{}, uint32(21), uint32(1), uint32(2), uint32(3), uint32(4), uint32(5), uint32(6))

	mock.ExpectQuery(`SELECT key,.* FROM expanded_compute_plans .* ORDER BY creation_date ASC, key ASC`).
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
	assert.Equal(t, uint32(21), plans[0].TaskCount)
	assert.Equal(t, uint32(1), plans[0].WaitingCount)
	assert.Equal(t, uint32(2), plans[0].TodoCount)
	assert.Equal(t, uint32(3), plans[0].DoingCount)
	assert.Equal(t, uint32(4), plans[0].CanceledCount)
	assert.Equal(t, uint32(5), plans[0].FailedCount)
	assert.Equal(t, uint32(6), plans[0].DoneCount)
	assert.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_FAILED, plans[0].Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryComputePlansNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"key", "owner", "delete_intermediary_models", "creation_date", "cancelation_date", "tag", "name", "metadata", "task_count", "waiting_count", "todo_count", "doing_count", "canceled_count", "failed_count", "done_count"})

	mock.ExpectQuery(`SELECT key,.* FROM expanded_compute_plans .* ORDER BY creation_date ASC, key`).
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
	defer mock.Close(context.Background())

	cpKey := "abc"
	cancelationDate, err := time.Parse("2006-01-02T15:04:05.000Z", "2021-02-03T04:05:06.007Z")
	assert.NoError(t, err)

	mock.ExpectBegin()

	mock.
		ExpectExec(`UPDATE compute_plans SET cancelation_date = $1 WHERE key = $2`).
		WithArgs(cancelationDate, cpKey).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.CancelComputePlan(&asset.ComputePlan{Key: cpKey}, cancelationDate)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
