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

package events

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/stretchr/testify/assert"
)

func TestEventHandling(t *testing.T) {
	eventSource := make(chan *fab.CCEvent)
	done := make(chan bool)

	events := 0

	listener := &Listener{
		events: eventSource,
		done:   done,
		onEvent: func(event *fab.CCEvent) {
			events++
		},
	}

	go listener.Listen()

	eventSource <- new(fab.CCEvent)
	eventSource <- new(fab.CCEvent)

	// We can't use listener.Close() since that requires an initialized contract
	done <- true

	assert.Equal(t, 2, events, "The callback should be called")
}
