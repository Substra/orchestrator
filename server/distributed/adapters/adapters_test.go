package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/owkin/orchestrator/server/common/interceptors"
	"github.com/stretchr/testify/assert"
)

func TestIsFabricTimeoutRetry(t *testing.T) {
	ctx := context.Background()
	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = interceptors.WithLastError(ctx, fmt.Errorf("test error"))
	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = interceptors.WithLastError(ctx, FabricTimeout)
	assert.True(t, isFabricTimeoutRetry(ctx))
}
