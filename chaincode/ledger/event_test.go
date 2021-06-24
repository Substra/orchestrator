package ledger

import (
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventDispatcher(t *testing.T) {
	stub := new(testHelper.MockedStub)
	stub.On("SetEvent", EventName, mock.Anything).Return(nil)

	dispatcher := newEventDispatcher(stub)

	err := dispatcher.Enqueue(&asset.Event{})
	assert.NoError(t, err)
	err = dispatcher.Enqueue(&asset.Event{})
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
