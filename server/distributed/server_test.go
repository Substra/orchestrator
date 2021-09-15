package distributed

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common"
)

func TestRetryOnSpecificError(t *testing.T) {
	assert.False(t, shouldRetry(fmt.Errorf("not an orchestration error")))
	assert.False(t, shouldRetry(errors.NewError(errors.ErrCannotDisableModel, "test")))
	assert.False(t, shouldRetry(errors.NewError(errors.ErrIncompatibleTaskStatus, "test")))
	assert.False(t, shouldRetry(errors.NewError(errors.ErrNotFound, "test")))

	assert.True(t, shouldRetry(fabricTimeout))
}

func TestIsFabricTimeoutRetry(t *testing.T) {
	ctx := context.Background()

	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = common.WithLastError(ctx, fmt.Errorf("test error"))

	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = common.WithLastError(ctx, fabricTimeout)

	assert.True(t, isFabricTimeoutRetry(ctx))
}
