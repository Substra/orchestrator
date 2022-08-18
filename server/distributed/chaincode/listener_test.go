package chaincode

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"google.golang.org/protobuf/encoding/protojson"
)

type extractTxIDTestCase struct {
	input    string
	valid    bool
	expected string
}

func TestExtractTxIDFromEventID(t *testing.T) {
	cases := map[string]extractTxIDTestCase{
		"valid": {
			"e00848bc-71c3-422f-b637-cbfc9d2e2042:10df4b99-d09e-4744-9093-6e98b9ec3bdb",
			true,
			"e00848bc-71c3-422f-b637-cbfc9d2e2042",
		},
		"invalid": {
			"foo",
			false,
			"",
		},
		"empty": {
			"",
			false,
			"",
		},
	}

	for name, tc := range cases {
		txID, err := extractTxIDFromEventID(tc.input)

		if tc.valid {
			assert.NoError(t, err, name+" should be valid")
			assert.Equal(t, tc.expected, txID)
		} else {
			assert.Error(t, err, name+" should not be valid")
		}
	}
}

type ccEventFactory struct {
	t *testing.T
}

func (f *ccEventFactory) newCCEvent(txID string, events ...*asset.Event) *fab.CCEvent {
	marshalledEvents := make([]json.RawMessage, len(events))
	for i, event := range events {
		marshalledEvent, err := protojson.Marshal(event)
		require.NoError(f.t, err)

		marshalledEvents[i] = marshalledEvent
	}

	payload, err := json.Marshal(marshalledEvents)
	require.NoError(f.t, err)

	return &fab.CCEvent{TxID: txID, Payload: payload}
}

func TestListenHandleCCEvent(t *testing.T) {
	ccEvents := make(chan *fab.CCEvent)

	nbCalls := 0

	listener := &Listener{
		ccEvents: ccEvents,
		handler: func(event *asset.Event) error {
			nbCalls++
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.TODO())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := listener.Listen(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	}()

	factory := ccEventFactory{t: t}
	ccEvents <- factory.newCCEvent("1", &asset.Event{Id: "1"}, &asset.Event{Id: "2"})

	cancel()
	wg.Wait()
	assert.Equal(t, 2, nbCalls)
}

func TestListenSkipCCEvent(t *testing.T) {
	ccEvents := make(chan *fab.CCEvent)

	nbCalls := 0

	listener := &Listener{
		ccEvents: ccEvents,
		handler: func(event *asset.Event) error {
			nbCalls++
			return nil
		},
		startTxID: "1",
	}

	ctx, cancel := context.WithCancel(context.TODO())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := listener.Listen(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	}()

	factory := ccEventFactory{t: t}
	ccEvents <- factory.newCCEvent("0")
	ccEvents <- factory.newCCEvent("1", &asset.Event{Id: "1"})

	cancel()
	wg.Wait()
	assert.Equal(t, 1, nbCalls)
}

func TestHandleCCEvent(t *testing.T) {
	nbCalls := 0

	listener := &Listener{
		handler: func(event *asset.Event) error {
			nbCalls++
			return nil
		},
	}

	factory := ccEventFactory{t: t}
	ccEvent := factory.newCCEvent("0", &asset.Event{Id: "2"})

	skipEvent, err := listener.handleCCEvent(ccEvent, false)
	assert.NoError(t, err)
	assert.False(t, skipEvent)
	assert.Equal(t, 1, nbCalls)
}

func TestHandleCCEventSkipEvent(t *testing.T) {
	nbCalls := 0
	listener := &Listener{
		handler: func(event *asset.Event) error {
			nbCalls++
			return nil
		},
		startEventID: "1",
	}

	factory := ccEventFactory{t: t}
	ccEvent := factory.newCCEvent(
		"0",
		&asset.Event{Id: "0"}, &asset.Event{Id: "1"}, &asset.Event{Id: "2"},
	)

	skipEvent, err := listener.handleCCEvent(ccEvent, true)
	assert.NoError(t, err)
	assert.False(t, skipEvent)
	assert.Equal(t, 1, nbCalls)
}
