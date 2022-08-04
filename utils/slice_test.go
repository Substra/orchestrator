package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceContains(t *testing.T) {
	haystack := []string{"test", "value", "end"}

	assert.True(t, SliceContains(haystack, "test"))
	assert.False(t, SliceContains(haystack, "not in slice"))
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

func TestUnique(t *testing.T) {
	input := []string{"item1", "item2", "item3", "item2", "item1"}
	expected := []string{"item1", "item2", "item3"}

	out := Unique(input)
	assert.Equal(t, expected, out)
}
