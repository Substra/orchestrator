package dbal

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/common"
)

func makeDataManagerRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"key", "name", "owner", "permissions", "description_address", "description_checksum", "opener_address", "opener_checksum", "type", "creation_date", "logs_permission", "metadata"}).
		AddRow("key1", "name", "owner", []byte("{}"), "https://example.com/desc", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "https://example.com/chksm", "993b6d90e0ed15d80e7e39c6fb298855d9544420be07faec52935649780e8f19", "", time.Unix(12, 0), []byte("{}"), map[string]string{}).
		AddRow("key2", "name", "owner", []byte("{}"), "https://example.com/desc", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "https://example.com/chksm", "993b6d90e0ed15d80e7e39c6fb298855d9544420be07faec52935649780e8f19", "", time.Unix(12, 0), []byte("{}"), map[string]string{})
}

func TestQueryDataManagers(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM expanded_datamanagers`).
		WithArgs(testChannel).
		WillReturnRows(makeDataManagerRows())

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, bookmark, err := dbal.QueryDataManagers(common.NewPagination("", 12))
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "", bookmark, "last page should be reached")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaginatedQueryDataManagers(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM expanded_datamanagers`).
		WithArgs(testChannel).
		WillReturnRows(makeDataManagerRows())

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, bookmark, err := dbal.QueryDataManagers(common.NewPagination("", 1))
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "1", bookmark, "There should be another page")

	assert.NoError(t, mock.ExpectationsWereMet())
}
