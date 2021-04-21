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
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/server/common"
)

// EventName is the name used by the orchestration chaincode to register its events on the ledger.
const EventName = "chaincode-updates"

// eventDispatcher is a struct storing events until their dispatch.
// Once contract processing is done, emitted events are aggregated into a single chaincode event
// and pushed with the transaction.
type eventDispatcher struct {
	event.Queue
	stub   shim.ChaincodeStubInterface
	logger log.Entry
}

// newEventDispatcher returns an eventDispatcher instance
func newEventDispatcher(stub shim.ChaincodeStubInterface) *eventDispatcher {
	return &eventDispatcher{
		Queue:  new(common.MemoryQueue),
		stub:   stub,
		logger: log.WithField("component", "event-dispatcher").WithField("mode", "chaincode"),
	}
}

// Dispatch aggregates events from the queue into a single chaincode event
// and assigns it to the transaction.
func (ed *eventDispatcher) Dispatch() error {
	if ed.Queue.Len() == 0 {
		ed.logger.Debug("No event to return with transaction")
		return nil
	}

	payload, err := json.Marshal(ed.Queue.GetEvents())
	if err != nil {
		return err
	}

	ed.logger.WithField("numBytes", len(payload)).Debug("Setting event to the transaction")

	err = ed.stub.SetEvent(EventName, payload)

	if err != nil {
		ed.logger.WithError(err).Error("Could not set event")
	}

	return err
}
