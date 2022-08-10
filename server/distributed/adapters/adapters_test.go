package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/server/common/interceptors"
)

func TestIsFabricTimeoutRetry(t *testing.T) {
	ctx := context.Background()
	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = interceptors.WithLastError(ctx, fmt.Errorf("test error"))
	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = interceptors.WithLastError(ctx, FabricTimeout)
	assert.True(t, isFabricTimeoutRetry(ctx))
}
