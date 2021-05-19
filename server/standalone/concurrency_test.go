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

package standalone

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrencyLimiter(t *testing.T) {
	limiter := NewConcurrencyLimiter()
	defer limiter.Stop()

	numTask := 5

	wg := new(sync.WaitGroup)
	wg.Add(numTask)

	start := make(chan struct{})
	done := make(chan int, numTask)

	for i := 0; i < numTask; i++ {
		go func(rank int) {
			token := <-limiter.AcquireExecutionToken()
			defer token.Release()
			if rank == 0 {
				// First task should wait for our go
				<-start
			}
			done <- rank
			wg.Done()
		}(i)

		// Wait for goroutine to start.
		// This rely on AcquireExecutionToken being called one after another, hence the delay
		// rather than a synchronization primitive (where goroutines may get interleaved).
		time.Sleep(50 * time.Millisecond)
	}

	assert.Len(t, done, 0, "no task should have been processed yet")

	// Start first task
	start <- struct{}{}

	// Wait for all task to finish
	wg.Wait()

	result := make([]int, 0, numTask)
	for len(done) > 0 {
		result = append(result, <-done)
	}

	prev := -1
	// result should be ordered
	for _, r := range result {
		assert.Greater(t, r, prev)
		prev = r
	}
}
