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

package chaincode

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/stretchr/testify/assert"
)

func TestContractInvocator(t *testing.T) {
	contract := &gateway.Contract{}

	invocator := NewContractInvocator(contract)

	assert.Implementsf(t, (*Invocator)(nil), invocator, "ContractInvocator should implements Invocator")
}
