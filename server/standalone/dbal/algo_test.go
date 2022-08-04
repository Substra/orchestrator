package dbal

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeAlgoRows(keys ...string) *pgxmock.Rows {
	permissions := []byte(`{"process": {"public": true}, "download": {"public": true}}`)

	res := pgxmock.NewRows([]string{"key", "name", "category", "description_address", "description_checksum", "algorithm_address", "algorithm_checksum", "permissions", "owner", "creation_date", "metadata"})

	for _, key := range keys {
		res.AddRow(key, "name", "ALGO_COMPOSITE", "address", "checksum", "address", "checksum", permissions, "owner", time.Unix(1337, 0), map[string]string{})
	}

	return res
}

func makeAlgoInputRows(algoKeys ...string) *pgxmock.Rows {
	res := pgxmock.NewRows([]string{"algoKey", "identifier", "kind", "multiple", "optional"})

	for _, algoKey := range algoKeys {
		res = res.AddRow(algoKey, "opener", asset.AssetKind_ASSET_DATA_MANAGER, true, false)
		res = res.AddRow(algoKey, "datasample", asset.AssetKind_ASSET_DATA_SAMPLE, true, false)
		res = res.AddRow(algoKey, "head", asset.AssetKind_ASSET_MODEL, false, true)
		res = res.AddRow(algoKey, "trunk", asset.AssetKind_ASSET_MODEL, false, true)
	}

	return res
}

func makeAlgoOutputRows(algoKeys ...string) *pgxmock.Rows {
	res := pgxmock.NewRows([]string{"algoKey", "identifier", "kind", "multiple"})

	for _, algoKey := range algoKeys {
		res = res.AddRow(algoKey, "head", asset.AssetKind_ASSET_MODEL, false)
		res = res.AddRow(algoKey, "trunk", asset.AssetKind_ASSET_MODEL, false)
	}

	return res
}

func TestQueryAlgos(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	computePlanKey := uuid.NewString()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT key, name, category, description_address, description_checksum, algorithm_address, algorithm_checksum, permissions, owner, creation_date, metadata FROM expanded_algos`).
		WithArgs(testChannel, computePlanKey).WillReturnRows(makeAlgoRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple, optional FROM algo_inputs WHERE algo_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeAlgoInputRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple FROM algo_outputs WHERE algo_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeAlgoOutputRows("key1", "key2"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	filter := &asset.AlgoQueryFilter{ComputePlanKey: computePlanKey}

	res, bookmark, err := dbal.QueryAlgos(common.NewPagination("", 12), filter)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Len(t, res[0].Inputs, 4)
	assert.Len(t, res[0].Outputs, 2)
	assert.Len(t, res[1].Inputs, 4)
	assert.Len(t, res[1].Outputs, 2)
	assert.Equal(t, "", bookmark, "last page should be reached")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaginatedQueryAlgos(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	computePlanKey := uuid.NewString()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT key, name, category, description_address, description_checksum, algorithm_address, algorithm_checksum, permissions, owner, creation_date, metadata FROM expanded_algos`).
		WithArgs(testChannel, computePlanKey).WillReturnRows(makeAlgoRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple, optional FROM algo_inputs WHERE algo_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeAlgoInputRows("key1"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple FROM algo_outputs WHERE algo_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeAlgoOutputRows("key1"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	filter := &asset.AlgoQueryFilter{ComputePlanKey: computePlanKey}

	res, bookmark, err := dbal.QueryAlgos(common.NewPagination("", 1), filter)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Len(t, res[0].Inputs, 4)
	assert.Len(t, res[0].Outputs, 2)
	assert.Equal(t, "1", bookmark, "There should be another page")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAlgo(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	uid := "key1"
	mock.ExpectQuery(`SELECT key, name, category, description_address, description_checksum, algorithm_address, algorithm_checksum, permissions, owner, creation_date, metadata FROM expanded_algos`).WillReturnRows(makeAlgoRows("key1"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple, optional FROM algo_inputs WHERE algo_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeAlgoInputRows("key1"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple FROM algo_outputs WHERE algo_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeAlgoOutputRows("key1"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, err := dbal.GetAlgo(uid)

	assert.NoError(t, err)
	assert.Len(t, res.Inputs, 4)
	assert.Len(t, res.Outputs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAlgoFail(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	uid := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`SELECT key, name, category, description_address, description_checksum, algorithm_address, algorithm_checksum, permissions, owner, creation_date, metadata FROM expanded_algos`).WillReturnError(pgx.ErrNoRows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, err = dbal.GetAlgo(uid)

	assert.Error(t, err)
	assert.ErrorContains(t, err, `not found`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryAlgosByComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT key, name, category, description_address, description_checksum, algorithm_address, algorithm_checksum, permissions, owner, creation_date, metadata FROM expanded_algos .* key IN \(SELECT DISTINCT`).
		WithArgs(testChannel, "CPKey").WillReturnRows(makeAlgoRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple, optional FROM algo_inputs WHERE algo_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeAlgoInputRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple FROM algo_outputs WHERE algo_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeAlgoOutputRows("key1", "key2"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, _, err := dbal.QueryAlgos(common.NewPagination("", 12), &asset.AlgoQueryFilter{ComputePlanKey: "CPKey"})
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryAlgosNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	mock.ExpectQuery(`key, name, category, description_address, description_checksum, algorithm_address, algorithm_checksum, permissions, owner, creation_date, metadata FROM expanded_algos`).
		WithArgs(testChannel).
		WillReturnRows(makeAlgoRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple, optional FROM algo_inputs WHERE algo_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeAlgoInputRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT algo_key, identifier, kind, multiple FROM algo_outputs WHERE algo_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeAlgoOutputRows("key1", "key2"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, _, err = dbal.QueryAlgos(common.NewPagination("", 12), nil)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
