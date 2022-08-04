package dbal

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataSampleFail(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	uid := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`SELECT key, owner, test_only, checksum, creation_date, datamanager_keys FROM expanded_datasamples`).
		WithArgs(testChannel, uid)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{
		tx:      tx,
		channel: testChannel,
	}

	_, err = dbal.GetDataSample(uid)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
