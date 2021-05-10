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

// Package testing contains helper functions to test components relying on orchestration events.
package testing

import (
	"reflect"

	"github.com/owkin/orchestrator/lib/event"
)

// EventMatcher returns a mock matcher function comparing the event received from call with the one expected.
// This matcher should be used instead of simple comparison to account for the generated ID.
// This function is to be used as a mock matcher, eg:
// queueMock.On("Enqueue", mock.MatchedBy(EventMatcher(&event.Event{AssetKind: asset.ModelKind})))
func EventMatcher(expected *event.Event) func(*event.Event) bool {
	return func(received *event.Event) bool {
		// Create a new event from the received one, without its generated ID
		withoutID := &event.Event{
			AssetKind: received.AssetKind,
			AssetKey:  received.AssetKey,
			EventKind: received.EventKind,
			Channel:   received.Channel,
			Metadata:  received.Metadata,
		}
		return reflect.DeepEqual(expected, withoutID)
	}
}
