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

package distributed

import (
	"encoding/json"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/contracts"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GatewayContract is the interface implemented by gateway.Contract.
// It should be in gateway module, but fabric rely on a concrete implementation rather than an interface...
type GatewayContract interface {
	Name() string
	EvaluateTransaction(name string, args ...string) ([]byte, error)
	SubmitTransaction(name string, args ...string) ([]byte, error)
	CreateTransaction(name string, opts ...gateway.TransactionOption) (*gateway.Transaction, error)
	RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error)
	Unregister(registration fab.Registration)
}

// Invocator describe how to invoke chaincode in a somewhat generic way.
// This is the Gandalf of fabric.
type Invocator interface {
	Call(method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error
}

// ContractInvocator implements the Invocator interface.
type ContractInvocator struct {
	contract      GatewayContract
	checker       contracts.TransactionChecker
	evaluatePeers []string
}

// NewContractInvocator creates an Invocator based on given smart contract.
func NewContractInvocator(c GatewayContract, checker contracts.TransactionChecker, evaluatePeers []string) *ContractInvocator {
	return &ContractInvocator{contract: c, checker: checker, evaluatePeers: evaluatePeers}
}

// Call will evaluate or invoke a transaction to the ledger, deserializing its result in the output parameter.
// The choice of evaluation or invocation is based on contracts.AllEvaluateTransactions.
func (i *ContractInvocator) Call(method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error {
	isEvaluate := i.checker.IsEvaluateMethod(method)

	txType := "Invoke"
	if isEvaluate {
		txType = "Evaluate"
	}

	logger := log.WithField("method", method).WithField("param", param).WithField("txType", txType)
	logger.Debug("Calling chaincode")

	start := time.Now()

	wrapper, err := communication.Wrap(param)
	if err != nil {
		return err
	}
	args, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	var data []byte

	if isEvaluate {
		var tx *gateway.Transaction
		tx, err = i.contract.CreateTransaction(method, gateway.WithEndorsingPeers(i.evaluatePeers...))

		if err != nil {
			return err
		}

		data, err = tx.Evaluate(string(args))
	} else {
		data, err = i.contract.SubmitTransaction(method, string(args))
	}

	if err != nil {
		return err
	}

	if output != nil {
		err := communication.Unwrap(data, output)
		if err != nil {
			return err
		}
	}

	elapsed := time.Since(start)

	logger.WithField("duration", elapsed).Debug("Successfully called chaincode")

	return nil
}
