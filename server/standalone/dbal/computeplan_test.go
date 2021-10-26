package dbal

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetComputeTasks(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset", "total", "waiting", "todo", "doing", "canceled", "failed", "done"}).
		AddRow(asset.ComputePlan{}, uint32(21), uint32(1), uint32(2), uint32(3), uint32(4), uint32(5), uint32(6))

	mock.ExpectQuery(`select cp.asset`).
		WithArgs("uuid", testChannel).
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPlanStatus(t *testing.T) {
	cases := map[string]struct {
		total    uint32
		done     uint32
		doing    uint32
		waiting  uint32
		failed   uint32
		canceled uint32
		outcome  asset.ComputePlanStatus
	}{
		"done": {
			total:    11,
			done:     11,
			doing:    0,
			waiting:  0,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_DONE,
		},
		"waiting": {
			total:    11,
			done:     0,
			doing:    0,
			waiting:  11,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_WAITING,
		},
		"failed": {
			total:    11,
			done:     1,
			doing:    0,
			waiting:  1,
			failed:   1,
			canceled: 1,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_FAILED,
		},
		"canceled": {
			total:    11,
			done:     1,
			doing:    0,
			waiting:  1,
			failed:   0,
			canceled: 1,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_CANCELED,
		},
		"doing": {
			total:    11,
			done:     1,
			doing:    0,
			waiting:  1,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_DOING,
		},
		"todo": {
			total:    11,
			done:     0,
			doing:    0,
			waiting:  10,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_TODO,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, getPlanStatus(tc.total, tc.done, tc.doing, tc.waiting, tc.failed, tc.canceled))
		})
	}
}

func TestGetRawComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset"}).
		AddRow(asset.ComputePlan{})

	mock.ExpectQuery(`select asset from compute_plans`).
		WithArgs("uuid", testChannel).
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	plan, err := dbal.GetRawComputePlan("uuid")
	assert.NoError(t, err)

	assert.Equal(t, uint32(0), plan.TaskCount)
	assert.Equal(t, uint32(0), plan.DoneCount)
	assert.Equal(t, asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN, plan.Status)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryComputePlans(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset", "total", "waiting", "todo", "doing", "canceled", "failed", "done"}).
		AddRow(asset.ComputePlan{}, uint32(21), uint32(1), uint32(2), uint32(3), uint32(4), uint32(5), uint32(6))

	mock.ExpectQuery(`select cp.asset,.* from "compute_plans" .* order by cp.asset->>'creationDate' asc, cp.id`).
		WithArgs(uint32(11), 0, testChannel).
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	plans, _, err := dbal.QueryComputePlans(common.NewPagination("", 10))
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
