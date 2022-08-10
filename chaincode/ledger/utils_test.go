package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
)

func TestGetTxCreator(t *testing.T) {
	org := "SampleOrg"

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, org), nil).Once()

	creator, err := GetTxCreator(stub)
	assert.Nil(t, err, "GetTxCreator should not fail")
	assert.Equal(t, org, creator, "Creator should be the MSP ID")
}
