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
func newEventDispatcher(stub shim.ChaincodeStubInterface, logger log.Entry) *eventDispatcher {
	return &eventDispatcher{
		Queue:  new(common.MemoryQueue),
		stub:   stub,
		logger: logger,
	}
}

// Dispatch aggregates events from the queue into a single chaincode event
// and assigns it to the transaction.
func (ed *eventDispatcher) Dispatch() error {
	if ed.Queue.Len() == 0 {
		ed.logger.Debug("No event to return with transaction")
		return nil
	}

	events := make([]json.RawMessage, ed.Queue.Len())
	for i, event := range ed.Queue.GetEvents() {
		b, err := marshaller.Marshal(event)
		if err != nil {
			return err
		}

		events[i] = b
	}

	payload, err := json.Marshal(events)
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
