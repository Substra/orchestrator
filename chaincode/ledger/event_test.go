package ledger

import (
	"testing"

	"github.com/go-playground/log/v7"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventDispatcher(t *testing.T) {
	stub := new(testHelper.MockedStub)
	stub.On("SetEvent", EventName, mock.Anything).Once().Return(nil)
	queue := new(common.MemoryQueue)

	dispatcher := newEventDispatcher(stub, queue, log.WithField("test", true))

	err := dispatcher.Enqueue(&asset.Event{})
	assert.NoError(t, err)
	err = dispatcher.Enqueue(&asset.Event{})
	assert.NoError(t, err)

	err = dispatcher.Dispatch()
	assert.NoError(t, err)

	stub.AssertExpectations(t)
}

func TestDispatchNoEvent(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(common.MemoryQueue)
	dispatcher := newEventDispatcher(stub, queue, log.WithField("test", true))

	// No "SetEvent" expected on stub: this will panic if SetEvent is called
	err := dispatcher.Dispatch()
	assert.NoError(t, err)
}
