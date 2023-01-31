package dbal

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
)

func makeFunctionRows(keys ...string) *pgxmock.Rows {
	permissions := []byte(`{"process": {"public": true}, "download": {"public": true}}`)

	res := pgxmock.NewRows([]string{"key", "name", "description_address", "description_checksum", "functionrithm_address", "functionrithm_checksum", "permissions", "owner", "creation_date", "metadata"})

	for _, key := range keys {
		res.AddRow(key, "name", "address", "checksum", "address", "checksum", permissions, "owner", time.Unix(1337, 0), map[string]string{})
	}

	return res
}

func makeFunctionInputRows(functionKeys ...string) *pgxmock.Rows {
	res := pgxmock.NewRows([]string{"functionKey", "identifier", "kind", "multiple", "optional"})

	for _, functionKey := range functionKeys {
		res = res.AddRow(functionKey, "opener", asset.AssetKind_ASSET_DATA_MANAGER, true, false)
		res = res.AddRow(functionKey, "datasample", asset.AssetKind_ASSET_DATA_SAMPLE, true, false)
		res = res.AddRow(functionKey, "head", asset.AssetKind_ASSET_MODEL, false, true)
		res = res.AddRow(functionKey, "trunk", asset.AssetKind_ASSET_MODEL, false, true)
	}

	return res
}

func makeFunctionOutputRows(functionKeys ...string) *pgxmock.Rows {
	res := pgxmock.NewRows([]string{"functionKey", "identifier", "kind", "multiple"})

	for _, functionKey := range functionKeys {
		res = res.AddRow(functionKey, "head", asset.AssetKind_ASSET_MODEL, false)
		res = res.AddRow(functionKey, "trunk", asset.AssetKind_ASSET_MODEL, false)
	}

	return res
}

func TestQueryFunctions(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	computePlanKey := uuid.NewString()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT key, name, description_address, description_checksum, functionrithm_address, functionrithm_checksum, permissions, owner, creation_date, metadata FROM expanded_functions`).
		WithArgs(testChannel, computePlanKey).WillReturnRows(makeFunctionRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple, optional FROM function_inputs WHERE function_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeFunctionInputRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple FROM function_outputs WHERE function_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeFunctionOutputRows("key1", "key2"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	filter := &asset.FunctionQueryFilter{ComputePlanKey: computePlanKey}

	res, bookmark, err := dbal.QueryFunctions(common.NewPagination("", 12), filter)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Len(t, res[0].Inputs, 4)
	assert.Len(t, res[0].Outputs, 2)
	assert.Len(t, res[1].Inputs, 4)
	assert.Len(t, res[1].Outputs, 2)
	assert.Equal(t, "", bookmark, "last page should be reached")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaginatedQueryFunctions(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	computePlanKey := uuid.NewString()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT key, name, description_address, description_checksum, functionrithm_address, functionrithm_checksum, permissions, owner, creation_date, metadata FROM expanded_functions`).
		WithArgs(testChannel, computePlanKey).WillReturnRows(makeFunctionRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple, optional FROM function_inputs WHERE function_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeFunctionInputRows("key1"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple FROM function_outputs WHERE function_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeFunctionOutputRows("key1"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	filter := &asset.FunctionQueryFilter{ComputePlanKey: computePlanKey}

	res, bookmark, err := dbal.QueryFunctions(common.NewPagination("", 1), filter)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Len(t, res[0].Inputs, 4)
	assert.Len(t, res[0].Outputs, 2)
	assert.Equal(t, "1", bookmark, "There should be another page")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFunction(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	uid := "key1"
	mock.ExpectQuery(`SELECT key, name, description_address, description_checksum, functionrithm_address, functionrithm_checksum, permissions, owner, creation_date, metadata FROM expanded_functions`).WillReturnRows(makeFunctionRows("key1"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple, optional FROM function_inputs WHERE function_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeFunctionInputRows("key1"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple FROM function_outputs WHERE function_key IN ($1)`)).
		WithArgs("key1").WillReturnRows(makeFunctionOutputRows("key1"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, err := dbal.GetFunction(uid)

	assert.NoError(t, err)
	assert.Len(t, res.Inputs, 4)
	assert.Len(t, res.Outputs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFunctionFail(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	uid := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`SELECT key, name, description_address, description_checksum, functionrithm_address, functionrithm_checksum, permissions, owner, creation_date, metadata FROM expanded_functions`).WillReturnError(pgx.ErrNoRows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, err = dbal.GetFunction(uid)

	assert.Error(t, err)
	assert.ErrorContains(t, err, `not found`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryFunctionsByComputePlan(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT key, name, description_address, description_checksum, functionrithm_address, functionrithm_checksum, permissions, owner, creation_date, metadata FROM expanded_functions .* key IN \(SELECT DISTINCT`).
		WithArgs(testChannel, "CPKey").WillReturnRows(makeFunctionRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple, optional FROM function_inputs WHERE function_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeFunctionInputRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple FROM function_outputs WHERE function_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeFunctionOutputRows("key1", "key2"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, _, err := dbal.QueryFunctions(common.NewPagination("", 12), &asset.FunctionQueryFilter{ComputePlanKey: "CPKey"})
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryFunctionsNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()

	mock.ExpectQuery(`key, name, description_address, description_checksum, functionrithm_address, functionrithm_checksum, permissions, owner, creation_date, metadata FROM expanded_functions`).
		WithArgs(testChannel).
		WillReturnRows(makeFunctionRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple, optional FROM function_inputs WHERE function_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeFunctionInputRows("key1", "key2"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT function_key, identifier, kind, multiple FROM function_outputs WHERE function_key IN ($1,$2)`)).
		WithArgs("key1", "key2").WillReturnRows(makeFunctionOutputRows("key1", "key2"))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, _, err = dbal.QueryFunctions(common.NewPagination("", 12), nil)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
