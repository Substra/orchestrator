package ledger

import (
	"testing"

	"github.com/go-playground/log/v7"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventDispatcher(t *testing.T) {
	stub := new(testHelper.MockedStub)
	stub.On("SetEvent", EventName, mock.Anything).Once().Return(nil)
	queue := new(event.MockQueue)

	queue.On("Len").Twice().Return(2)
	queue.On("GetEvents").Once().Return([]*asset.Event{{}, {}})

	dispatcher := newEventDispatcher(stub, queue, log.WithField("test", true))

	err := dispatcher.Dispatch()
	assert.NoError(t, err)

	stub.AssertExpectations(t)
	queue.AssertExpectations(t)
}

func TestDispatchNoEvent(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(event.MockQueue)
	dispatcher := newEventDispatcher(stub, queue, log.WithField("test", true))

	queue.On("Len").Return(0)

	// No "SetEvent" expected on stub: this will panic if SetEvent is called
	err := dispatcher.Dispatch()
	assert.NoError(t, err)

	stub.AssertExpectations(t)
	queue.AssertExpectations(t)
}
