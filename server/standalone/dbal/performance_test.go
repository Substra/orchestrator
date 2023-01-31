package dbal

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
)

func TestPerformanceNotFound(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	taskKey := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	metricKey := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	rows := pgxmock.NewRows([]string{"compute_task_key", "function_key", "performance_value", "creation_date"})
	mock.ExpectQuery(`SELECT compute_task_key, function_key, performance_value, creation_date FROM performances`).
		WithArgs(testChannel, taskKey, metricKey).
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	pagination := common.NewPagination("", 100)
	performances, _, err := dbal.QueryPerformances(pagination, &asset.PerformanceQueryFilter{
		ComputeTaskKey: taskKey,
		MetricKey:      metricKey,
	})
	assert.Len(t, performances, 0)
	assert.NoError(t, err)

}

func TestQueryPerformancesNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"compute_task_key", "function_key", "performance_value", "creation_date"})
	mock.ExpectQuery(`SELECT compute_task_key, function_key, performance_value, creation_date FROM performances`).
		WithArgs(testChannel).
		WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	pagination := common.NewPagination("", 100)
	_, _, err = dbal.QueryPerformances(pagination, nil)
	assert.NoError(t, err)
}
