package ledger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
)

func TestCountComputeTaskRegisteredOutputs(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(context.TODO(), stub, queue)

	stub.On("GetState", "computetask_output_asset:test").Return([]byte{}, nil).Twice()

	counter, err := db.CountComputeTaskRegisteredOutputs("test")

	assert.NoError(t, err)
	assert.Equal(t, len(counter), 0)
}
