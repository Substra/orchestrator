package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("underlying errror")

func TestMessageFormating(t *testing.T) {
	err := NewError(ErrNotFound, "asset not found")
	assert.Equal(t, "OE0002: asset not found", err.Error())

	err = err.Wrap(fmt.Errorf("another error"))
	assert.Equal(t, "OE0002: asset not found: another error", err.Error())
}

func TestErrorWrapping(t *testing.T) {
	err := NewError(ErrInternalError, "test").Wrap(errTest)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, errTest))
	assert.EqualError(t, errors.Unwrap(err), errTest.Error())

	outErr := new(OrcError)
	assert.True(t, errors.As(err, &outErr))

	assert.Equal(t, ErrInternalError, outErr.Kind)
}
