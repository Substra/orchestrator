package testing

import (
	"testing"

	// fabric-protos-go/msp still rely on this deprecated lib
	"github.com/golang/protobuf/proto" // nolint:staticcheck
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/stretchr/testify/require"
)

// FakeTxCreator handles the dirty details of generating a transaction creator
func FakeTxCreator(t *testing.T, mspid string) []byte {
	sID := msp.SerializedIdentity{
		Mspid: mspid,
	}
	b, err := proto.Marshal(&sID)
	require.Nil(t, err, "SID marshal should not fail")

	return b
}
