package standalone

import (
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/errors"
)

func TestRetryOnUnserializableTransaction(t *testing.T) {
	assert.True(t, shouldRetry(&pgconn.PgError{Code: "40001"}))
	assert.True(t, shouldRetry(errors.NewError(errors.ErrNotFound, "test").Wrap(&pgconn.PgError{Code: "40001"})))
	assert.False(t, shouldRetry(&pgconn.PgError{Code: "1234"}))
	assert.False(t, shouldRetry(fmt.Errorf("not a pgconn error")))
}
