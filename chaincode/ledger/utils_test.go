package ledger

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testHelper "github.com/substrafoundation/substra-orchestrator/chaincode/testing"
)

func TestGetTxCreator(t *testing.T) {
	org := "SampleOrg"

	sID := msp.SerializedIdentity{
		Mspid: org,
	}

	b, err := proto.Marshal(&sID)
	require.Nil(t, err, "SID marshal should not fail")

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	creator, err := GetTxCreator(stub)
	assert.Nil(t, err, "GetTxCreator should not fail")
	assert.Equal(t, org, creator, "Creator should be the MSP ID")
}
