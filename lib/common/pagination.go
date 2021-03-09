// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
