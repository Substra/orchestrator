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
