package distributed

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/owkin/orchestrator/lib/errors"
)

func TestRetryOnSpecificError(t *testing.T) {
	assert.False(t, shouldRetry(fmt.Errorf("not an orchestration error")))
	assert.False(t, shouldRetry(fmt.Errorf("%w", errors.ErrCannotDisableModel)))
	assert.True(t, shouldRetry(fmt.Errorf("%w", errors.ErrIncompatibleTaskStatus)))
	assert.True(t, shouldRetry(fmt.Errorf("%w", errors.ErrNotFound)))
	assert.True(t, shouldRetry(fmt.Errorf("%w", errors.ErrReferenceNotFound)))
}
