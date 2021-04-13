// Copyright 2020 Owkin Inc.
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

// Package utils contains various small utility functions
package utils

import (
	"reflect"
	"sort"
)

// StringInSlice will check if needle is found in haystack
func StringInSlice(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// Combine combines two lists of string in a single one without duplicates.
func Combine(list1 []string, list2 []string) []string {
	return append(list1, Filter(list2, list1)...)
}

// Returns the content of list 1 not present in list 2.
// list1 - list1 U list2
func Filter(list1 []string, list2 []string) []string {
	var output []string
	for _, item := range list1 {
		ok := StringInSlice(list2, item)
		if !ok {
			output = append(output, item)
		}
	}
	return output

}

// IsEqual compares two slices and returns true if they both contains the same set of items.
func IsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))

	copy(aCopy, a)
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	return reflect.DeepEqual(aCopy, bCopy)
}

// Intersection returns a new slice containing items in common from a and b
func Intersection(a, b []string) []string {
	res := []string{}
	for _, n := range a {
		if StringInSlice(b, n) {
			res = append(res, n)
		}
	}
	return res
}
