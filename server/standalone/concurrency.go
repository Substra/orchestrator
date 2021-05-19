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
	"container/list"
	"sync"
)

// ExecutionToken is a token given by a RequestScheduler.
// Once a routine acquired an ExecutionToken, it can be processed.
// The token should be released once the routine has completed.
type ExecutionToken struct {
	done chan<- struct{}
}

func (e *ExecutionToken) Release() {
	e.done <- struct{}{}
}

// RequestScheduler is a synchronization mechanism.
// All scheduled routines should share the same RequestScheduler and
// wait for an available token by calling AcquireToken().
// Once a routine has a token, other routine won't be issued one
// until the running routine releases its token with token.Release().
type RequestScheduler interface {
	AcquireExecutionToken() <-chan ExecutionToken
}

type job struct {
	tokenChannel chan<- ExecutionToken
}

// ConcurrencyLimiter implements RequestScheduler interface and
// limit concurrency to a single routine.
// It is used in standalone mode to make sure only one request can
// go through the orchestrator at a time.
// Order of requests is preserved (FIFO pattern).
type ConcurrencyLimiter struct {
	fifo   *list.List
	m      *sync.RWMutex
	hasJob chan struct{}
	stop   chan struct{}
}

// NewConcurrencyLimiter returns a RequestScheduler instance ready to use.
func NewConcurrencyLimiter() *ConcurrencyLimiter {
	scheduler := &ConcurrencyLimiter{
		// FIFO queue of pending jobs, from back to front
		fifo:   list.New(),
		m:      new(sync.RWMutex),
		hasJob: make(chan struct{}),
		stop:   make(chan struct{}),
	}

	go scheduler.schedule()

	return scheduler
}

func (c *ConcurrencyLimiter) AcquireExecutionToken() <-chan ExecutionToken {
	out := make(chan ExecutionToken)

	j := job{
		tokenChannel: out,
	}

	go c.addJob(j)

	return out
}

func (c *ConcurrencyLimiter) Stop() {
	c.stop <- struct{}{}
}

func (c *ConcurrencyLimiter) addJob(j job) {
	c.m.Lock()
	c.fifo.PushBack(j)
	c.m.Unlock()
	c.hasJob <- struct{}{}
}

// schedule is the main loop of the scheduler, where jobs are issued tokens
// in the order in which they were enqueued.
func (c *ConcurrencyLimiter) schedule() {
	for {
		select {
		case <-c.hasJob:
			c.m.RLock()
			e := c.fifo.Front()
			c.m.RUnlock()
			j := e.Value.(job)

			release := make(chan struct{})
			token := ExecutionToken{done: release}

			// send a token to start the job
			go func() {
				j.tokenChannel <- token
			}()

			// wait for token to be released
			<-release

			c.m.Lock()
			c.fifo.Remove(e)
			c.m.Unlock()
		case <-c.stop:
			return
		}
	}
}

// ImmediateRequestScheduler is a RequestScheduler which does not block.
// This is useful when testing handlers.
type ImmediateRequestScheduler struct{}

func (s *ImmediateRequestScheduler) AcquireExecutionToken() <-chan ExecutionToken {
	out := make(chan ExecutionToken)

	go func() {
		release := make(chan struct{})
		out <- ExecutionToken{done: release}
		<-release
	}()
	return out
}
