package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringInSlice(t *testing.T) {
	haystack := []string{"test", "value", "end"}

	assert.True(t, StringInSlice(haystack, "test"))
	assert.False(t, StringInSlice(haystack, "not in slice"))
}

func TestCombine(t *testing.T) {
	list1 := []string{"item1", "item2", "item3"}
	list2 := []string{"item2", "item4"}

	out := Combine(list1, list2)
	assert.Equal(t, out, []string{"item1", "item2", "item3", "item4"})
}

func TestFilter(t *testing.T) {
	list1 := []string{"item1", "item2", "item3"}
	list2 := []string{"item2", "item4"}

	out := Filter(list1, list2)
	assert.Equal(t, out, []string{"item1", "item3"})
}
