// Package utils contains various small utility functions
package utils

import (
	"reflect"
	"sort"
)

// SliceContains will check if needle is found in haystack
func SliceContains[T comparable](haystack []T, needle T) bool {
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

// Filter returns the content of list 1 not present in list 2.
// list1 - list1 U list2
func Filter(list1 []string, list2 []string) []string {
	var output []string
	for _, item := range list1 {
		ok := SliceContains(list2, item)
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
		if SliceContains(b, n) {
			res = append(res, n)
		}
	}
	return res
}

// Unique returns a slice without duplicates.
func Unique[T comparable](arr []T) []T {
	keys := make(map[T]bool)
	list := make([]T, 0, len(arr))
	for _, entry := range arr {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
