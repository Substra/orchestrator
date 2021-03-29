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
