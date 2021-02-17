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

package common

import "github.com/owkin/orchestrator/lib/event"

// MemoryQueue keeps events in memory
type MemoryQueue struct {
	events []*event.Event
}

// Enqueue adds an event to the queue
func (q *MemoryQueue) Enqueue(event *event.Event) error {
	q.events = append(q.events, event)

	return nil
}

// GetEvents returns queued events
func (q *MemoryQueue) GetEvents() []*event.Event {
	return q.events
}

// Len returns the length of the queue
func (q *MemoryQueue) Len() int {
	return len(q.events)
}
