package ledger

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
)

func TestAddExistingAlgo(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)

	db := NewDB(context.TODO(), stub, queue)

	algo := &asset.Algo{Key: "test"}

	stub.On("GetState", "algo:test").Return([]byte("{}"), nil).Once()

	err := db.AddAlgo(algo)
	orcErr := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcErr))
	assert.Equal(t, orcerrors.ErrConflict, orcErr.Kind)
}
