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
	"encoding/json"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Invocator describe how to invoke chaincode in a somewhat generic way.
// This is the Gandalf of fabric.
type Invocator interface {
	Invoke(method string, args []string, output interface{}) error
}

// ContractInvocator implements the Invocator interface.
type ContractInvocator struct {
	contract *gateway.Contract
}

// NewContractInvocator creates an Invocator based on given smart contract.
func NewContractInvocator(c *gateway.Contract) *ContractInvocator {
	return &ContractInvocator{c}
}

// Invoke will submit a transaction to the ledger, deserializing its result in the output parameter.
func (i *ContractInvocator) Invoke(method string, args []string, output interface{}) error {
	logger := log.WithField("method", method)

	logger.WithField("args", args).Debug("Invoking chaincode")
	start := time.Now()

	data, err := i.contract.SubmitTransaction(method, args...)

	if err != nil {
		logger.WithError(err).Error("Failed to invoke chaincode")
		return err
	}

	err = json.Unmarshal(data, output)

	if err != nil {
		logger.WithError(err).WithField("data", data).Error("Failed to deserialize")
		return err
	}

	elapsed := time.Since(start)

	logger.WithField("duration", elapsed).Debug("Invokation successful")

	return nil
}
