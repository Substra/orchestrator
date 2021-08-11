package distributed

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/stretchr/testify/assert"

	"github.com/owkin/orchestrator/lib/errors"
)

func TestRetryOnSpecificError(t *testing.T) {
	assert.False(t, shouldRetry(fmt.Errorf("not an orchestration error")))
	assert.False(t, shouldRetry(fmt.Errorf("%w", errors.ErrCannotDisableModel)))
	assert.True(t, shouldRetry(fmt.Errorf("%w", errors.ErrIncompatibleTaskStatus)))
	assert.True(t, shouldRetry(fmt.Errorf("%w", errors.ErrNotFound)))
	assert.True(t, shouldRetry(fmt.Errorf("%w", errors.ErrReferenceNotFound)))

	fabricTimeout := status.New(status.ClientStatus, status.Timeout.ToInt32(), "request timed out or been cancelled", nil)
	assert.True(t, shouldRetry(fabricTimeout))
}
