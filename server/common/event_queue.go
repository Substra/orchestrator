package common

import "github.com/substra/orchestrator/lib/asset"

// MemoryQueue keeps events in memory
type MemoryQueue struct {
	events []*asset.Event
}

// Enqueue adds an event to the queue
func (q *MemoryQueue) Enqueue(event *asset.Event) error {
	q.events = append(q.events, event)

	return nil
}

// GetEvents returns queued events
func (q *MemoryQueue) GetEvents() []*asset.Event {
	return q.events
}

// Len returns the length of the queue
func (q *MemoryQueue) Len() int {
	return len(q.events)
}
