package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCaller(t *testing.T) {
	caller, err := GetCaller(0)
	assert.NoError(t, err)
	assert.Equal(t, "TestGetCaller", caller)

	caller, err = myFunction()
	assert.NoError(t, err)
	assert.Equal(t, "TestGetCaller", caller)
}

func myFunction() (string, error) {
	return GetCaller(1)
}
