package dbal

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v4"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFailureReportNotFound(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	mock.ExpectBegin()

	computeTaskKey := "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"
	mock.ExpectQuery(`SELECT .* FROM expanded_failure_reports`).
		WithArgs(testChannel, computeTaskKey).
		WillReturnError(pgx.ErrNoRows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, err = dbal.GetFailureReport(computeTaskKey)

	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrNotFound, orcError.Kind)
	assert.NoError(t, mock.ExpectationsWereMet())
}
