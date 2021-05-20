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

// Package event defines orchestration Event structure and interfaces to interact with it.
// Events are emitted by orchestration logic and are placed in a Queue.
// Once the orchestration action is done, those events can be sent to a broker through a Dispatcher.
package event

import "github.com/owkin/orchestrator/lib/asset"

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
	Queue
	Dispatch() error
}

// QueueProvider defines an object able to provide a Queue instance
type QueueProvider interface {
	GetEventQueue() Queue
}
