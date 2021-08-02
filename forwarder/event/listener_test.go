package event

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/stretchr/testify/assert"
)

func TestEventHandling(t *testing.T) {
	eventSource := make(chan *fab.CCEvent)
	done := make(chan bool)

	eventIdx := new(MockIndexer)

	eventIdx.On("GetLastEvent", "testChannel").Once().Return(IndexedEvent{})

	events := 0

	listener := &Listener{
		events: eventSource,
		done:   done,
		handler: func(event *fab.CCEvent) {
			events++
		},
		eventIdx: eventIdx,
		channel:  "testChannel",
	}

	go listener.Listen()

	event1 := &fab.CCEvent{
		TxID: "1",
	}
	event2 := &fab.CCEvent{
		TxID: "2",
	}

	eventIdx.On("SetLastEvent", "testChannel", event1).Return(nil)
	eventIdx.On("SetLastEvent", "testChannel", event2).Return(nil)

	eventSource <- event1
	eventSource <- event2

	// We can't use listener.Close() since that requires an initialized contract
	done <- true

	assert.Equal(t, 2, events, "The callback should be called")
}

func TestEventSkipping(t *testing.T) {
	eventSource := make(chan *fab.CCEvent)
	done := make(chan bool)

	eventIdx := new(MockIndexer)

	eventIdx.On("GetLastEvent", "testChannel").Once().Return(IndexedEvent{BlockNum: 12, TxID: "lastKnown"})

	events := 0

	listener := &Listener{
		events: eventSource,
		done:   done,
		handler: func(event *fab.CCEvent) {
			if event.TxID == "4" {
				events++
			}
		},
		eventIdx: eventIdx,
		channel:  "testChannel",
	}

	go listener.Listen()

	// events 1-3 should be skipped
	event1 := &fab.CCEvent{
		TxID: "1",
	}
	event2 := &fab.CCEvent{
		TxID: "2",
	}
	event3 := &fab.CCEvent{
		TxID: "lastKnown",
	}
	event4 := &fab.CCEvent{
		TxID: "4",
	}

	eventIdx.On("SetLastEvent", "testChannel", event4).Return(nil)

	eventSource <- event1
	eventSource <- event2
	eventSource <- event3
	eventSource <- event4

	// We can't use listener.Close() since that requires an initialized contract
	done <- true

	assert.Equal(t, 1, events, "The callback should be called only once")
}
