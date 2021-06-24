package communication

import (
	"encoding/json"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestWrapUnwrap(t *testing.T) {
	msg := &asset.NewAlgo{
		Key:      "uuid",
		Category: asset.AlgoCategory_ALGO_SIMPLE,
	}

	wrapped, err := Wrap(msg)
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
