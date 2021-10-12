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

func TestPerformanceNotFound(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	defer mock.Close(context.Background())

	mock.ExpectBegin()

	taskKey := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	metricKey := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	rows := pgxmock.NewRows([]string{"asset"})
	mock.ExpectQuery(`SELECT asset FROM performances WHERE channel = \$1 AND compute_task_id = \$2 AND metric_id = \$3 ORDER BY asset->>'creationDate' ASC, metric_id DESC, compute_task_id DESC LIMIT 101 OFFSET 0`).WithArgs(testChannel, taskKey, metricKey).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	pagination := common.NewPagination("", 100)
	performances, _, err := dbal.QueryPerformances(pagination, &asset.PerformanceQueryFilter{
		ComputeTaskKey: taskKey,
		MetricKey:      metricKey,
	})
	assert.Equal(t, len(performances), 0)
	assert.Equal(t, err, nil)

}
