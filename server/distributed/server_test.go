package distributed

import (
	"fmt"
	"testing"

	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/distributed/adapters"
	"github.com/stretchr/testify/assert"
)

func TestRetryOnSpecificError(t *testing.T) {
	assert.False(t, shouldRetry(fmt.Errorf("not an orchestration error")))
	assert.False(t, shouldRetry(errors.NewError(errors.ErrCannotDisableModel, "test")))
	assert.False(t, shouldRetry(errors.NewError(errors.ErrIncompatibleTaskStatus, "test")))
	assert.False(t, shouldRetry(errors.NewError(errors.ErrNotFound, "test")))

	assert.True(t, shouldRetry(adapters.FabricTimeout))
}
