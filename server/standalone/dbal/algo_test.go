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

func TestQueryAlgos(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT asset FROM algos`).WithArgs(testChannel, asset.AlgoCategory_ALGO_COMPOSITE.String()).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, bookmark, err := dbal.QueryAlgos(common.NewPagination("", 12), &asset.AlgoQueryFilter{Category: asset.AlgoCategory_ALGO_COMPOSITE})
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "", bookmark, "last page should be reached")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPaginatedQueryAlgos(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT asset FROM algos`).WithArgs(testChannel, asset.AlgoCategory_ALGO_COMPOSITE.String()).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, bookmark, err := dbal.QueryAlgos(common.NewPagination("", 1), &asset.AlgoQueryFilter{Category: asset.AlgoCategory_ALGO_COMPOSITE})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "1", bookmark, "There should be another page")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetAlgoFail(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	defer mock.Close(context.Background())

	mock.ExpectBegin()

	uid := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`select "asset" from "algos" where id=`).WithArgs(uid, testChannel)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, err = dbal.GetAlgo(uid)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryAlgosByComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT asset FROM algos .* id IN \(SELECT DISTINCT`).WithArgs(testChannel, "CPKey").WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, _, err := dbal.QueryAlgos(common.NewPagination("", 12), &asset.AlgoQueryFilter{ComputePlanKey: "CPKey"})
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
