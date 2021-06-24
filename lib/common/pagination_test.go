// Package common contains generic entities used by other modules of the library.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPagination(t *testing.T) {
	paginationWithoutSize := NewPagination("", 0)
	assert.Equal(t, uint32(DefaultPageSize), paginationWithoutSize.Size)

	paginationWithSize := NewPagination("uuid", 12)
	assert.Equal(t, uint32(12), paginationWithSize.Size)
	assert.Equal(t, "uuid", paginationWithSize.Token)
}
