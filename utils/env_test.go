package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenvBool(t *testing.T) {
	_, err := GetenvBool("UNSET_TEST_ORCHESTRATOR_VALUE___")
	assert.Error(t, err, "Value should not exist")

	os.Setenv("TEST_ORCHESTRATOR_VALUE_TRUE___", "true")
	val, err := GetenvBool("TEST_ORCHESTRATOR_VALUE_TRUE___")
	assert.NoError(t, err)
	assert.True(t, val)

	os.Setenv("TEST_ORCHESTRATOR_VALUE_FALSE___", "false")
	val, err = GetenvBool("TEST_ORCHESTRATOR_VALUE_FALSE___")
	assert.NoError(t, err)
	assert.False(t, val)
}
