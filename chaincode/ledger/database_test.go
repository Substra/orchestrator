package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFullKey(t *testing.T) {
	k := getFullKey("resource", "id")

	assert.Equal(t, "resource:id", k, "key should be prefixed with resource type")
}
