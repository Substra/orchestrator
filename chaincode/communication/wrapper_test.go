package communication

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestWrapUnwrap(t *testing.T) {
	msg := &asset.NewAlgo{Key: "uuid"}

	wrapped, err := Wrap(context.Background(), msg)
	assert.NoError(t, err)

	out := new(asset.NewAlgo)
	err = wrapped.Unwrap(out)
	assert.NoError(t, err)
	assert.Equal(t, msg, out)

	serialized, err := json.Marshal(wrapped)
	assert.NoError(t, err)

	out = new(asset.NewAlgo)
	err = Unwrap(serialized, out)
	assert.NoError(t, err)
	assert.Equal(t, msg, out)
}
