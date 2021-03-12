// Copyright 2020 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
