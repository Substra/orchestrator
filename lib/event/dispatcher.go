// Package event defines orchestration Event structure and interfaces to interact with it.
// Events are emitted by orchestration logic and are placed in a Queue.
// Once the orchestration action is done, those events can be sent to a broker through a Dispatcher.
package event

import "github.com/substra/orchestrator/lib/asset"

// Queue holds events while the transaction is being processed.
// Events are eventually dispatched by a Dispatcher once processing is done.
type Queue interface {
	Enqueue(event *asset.Event) error
	GetEvents() []*asset.Event
	Len() int
}

// Dispatcher is responsible for broadcasting events
// Events are added via Push and dispatched with Dispatch
type Dispatcher interface {
	Dispatch() error
}

// QueueProvider defines an object able to provide a Queue instance
type QueueProvider interface {
	GetEventQueue() Queue
}
