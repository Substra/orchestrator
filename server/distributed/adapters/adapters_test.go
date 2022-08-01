package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/owkin/orchestrator/server/common"
	"github.com/stretchr/testify/assert"
)

func TestIsFabricTimeoutRetry(t *testing.T) {
	ctx := context.Background()
	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = common.WithLastError(ctx, fmt.Errorf("test error"))
	assert.False(t, isFabricTimeoutRetry(ctx))

	ctx = common.WithLastError(ctx, FabricTimeout)
	assert.True(t, isFabricTimeoutRetry(ctx))
}
