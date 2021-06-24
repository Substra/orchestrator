package ledger

import (
	// fabric-protos-go/msp still rely on this deprecated lib
	"github.com/golang/protobuf/proto" // nolint:staticcheck
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// GetTxCreator returns the transaction creator
func GetTxCreator(stub shim.ChaincodeStubInterface) (string, error) {
	creator, err := stub.GetCreator()

	if err != nil {
		return "", err
	}

	sID := &msp.SerializedIdentity{}

	err = proto.Unmarshal(creator, sID)
	if err != nil {
		return "", err
	}

	return sID.GetMspid(), nil
}
