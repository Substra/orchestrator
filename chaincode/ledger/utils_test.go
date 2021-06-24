package ledger

import (
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/stretchr/testify/assert"
)

func TestGetTxCreator(t *testing.T) {
	org := "SampleOrg"

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, org), nil).Once()

	creator, err := GetTxCreator(stub)
	assert.Nil(t, err, "GetTxCreator should not fail")
	assert.Equal(t, org, creator, "Creator should be the MSP ID")
}
