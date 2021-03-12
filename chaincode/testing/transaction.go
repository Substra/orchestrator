// Copyright 2021 Owkin Inc.
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
