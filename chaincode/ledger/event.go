package ledger

import (
	"encoding/json"

	"github.com/substra/orchestrator/lib/asset"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/rs/zerolog"
)

// EventName is the name used by the orchestration chaincode to register its events on the ledger.
const EventName = "chaincode-updates"

// EventQueue holds events while the transaction is being processed.
// Events are eventually dispatched by an EventDispatcher once processing is done.
type EventQueue interface {
	Enqueue(event *asset.Event) error
	GetEvents() []*asset.Event
	Len() int
}

// EventDispatcher is responsible for pushing events with the transaction.
type EventDispatcher interface {
	Dispatch() error
}

// eventDispatcher is a struct storing events until their dispatch.
// Once contract processing is done, emitted events are aggregated into a single chaincode event
// and pushed with the transaction.
type eventDispatcher struct {
	queue  EventQueue
	stub   shim.ChaincodeStubInterface
	logger *zerolog.Logger
}

// newEventDispatcher returns an eventDispatcher instance
func newEventDispatcher(stub shim.ChaincodeStubInterface, queue EventQueue, logger *zerolog.Logger) *eventDispatcher {
	return &eventDispatcher{
		queue:  queue,
		stub:   stub,
		logger: logger,
	}
}

// Dispatch aggregates events from the queue into a single chaincode event
// and assigns it to the transaction.
func (ed *eventDispatcher) Dispatch() error {
	if ed.queue.Len() == 0 {
		ed.logger.Debug().Msg("No event to return with transaction")
		return nil
	}

	events := make([]json.RawMessage, ed.queue.Len())
	for i, event := range ed.queue.GetEvents() {
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

	ed.logger.Debug().Int("numBytes", len(payload)).Msg("Setting event to the transaction")

	err = ed.stub.SetEvent(EventName, payload)

	if err != nil {
		ed.logger.Error().Err(err).Msg("Could not set event")
	}

	return err
}
