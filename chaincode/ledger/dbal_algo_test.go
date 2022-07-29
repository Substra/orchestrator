package ledger

import (
	"context"
	"errors"
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/stretchr/testify/assert"
)

func TestAddExistingAlgo(t *testing.T) {
	stub := new(testHelper.MockedStub)

	db := NewDB(context.TODO(), stub)

	algo := &asset.Algo{Key: "test"}

	stub.On("GetState", "algo:test").Return([]byte("{}"), nil).Once()

	err := db.AddAlgo(algo)
	orcErr := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcErr))
	assert.Equal(t, orcerrors.ErrConflict, orcErr.Kind)
}
