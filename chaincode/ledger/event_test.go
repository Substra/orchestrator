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

package ledger

import (
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventDispatcher(t *testing.T) {
	stub := new(testHelper.MockedStub)
	stub.On("SetEvent", EventName, mock.Anything).Return(nil)

	dispatcher := newEventDispatcher(stub)

	err := dispatcher.Enqueue(&event.Event{})
	assert.NoError(t, err)
	err = dispatcher.Enqueue(&event.Event{})
	assert.NoError(t, err)

	err = dispatcher.Dispatch()
	assert.NoError(t, err)
}

func TestDispatchNoEvent(t *testing.T) {
	stub := new(testHelper.MockedStub)
	dispatcher := newEventDispatcher(stub)

	// No "SetEvent" expected on stub: this will panic if SetEvent is called
	err := dispatcher.Dispatch()
	assert.NoError(t, err)
}
