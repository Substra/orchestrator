package standalone

import (
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/stretchr/testify/assert"
)

func TestRetryOnUnserializableTransaction(t *testing.T) {
	assert.True(t, shouldRetry(&pgconn.PgError{Code: "40001"}))
	assert.True(t, shouldRetry(errors.NewError(errors.ErrNotFound, "test").Wrap(&pgconn.PgError{Code: "40001"})))
	assert.False(t, shouldRetry(&pgconn.PgError{Code: "1234"}))
	assert.False(t, shouldRetry(fmt.Errorf("not a pgconn error")))
}
