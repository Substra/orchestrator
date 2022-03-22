package dbal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testChannel = "testchannel"

func TestGetOffset(t *testing.T) {
	emptyOffset, err := getOffset("")
	assert.NoError(t, err)
	assert.Equal(t, 0, emptyOffset, "empty token should default to zero")

	valueOffset, err := getOffset("12")
	assert.NoError(t, err)
	assert.Equal(t, 12, valueOffset, "valued token should be parserd as int")
}
