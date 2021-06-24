// Package common contains generic entities used by other modules of the library.
package common

// DefaultPageSize is the size used for pagination when not explicitly set by caller
const DefaultPageSize = 50

// PaginationToken is the starting point of paginated lists.
type PaginationToken = string

// Pagination represents a pagination request
type Pagination struct {
	Token PaginationToken
	Size  uint32
}

// NewPagination returns a new Pagination object.
// If size is null, it will defaults to DefaultPageSize.
func NewPagination(token string, size uint32) *Pagination {
	pageSize := size
	if size <= 0 {
		pageSize = DefaultPageSize
	}

	return &Pagination{Token: token, Size: pageSize}
}
