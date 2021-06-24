package event

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
