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

package contracts

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/algo"
	"github.com/owkin/orchestrator/chaincode/computetask"
	"github.com/owkin/orchestrator/chaincode/datamanager"
	"github.com/owkin/orchestrator/chaincode/datasample"
	"github.com/owkin/orchestrator/chaincode/node"
	"github.com/owkin/orchestrator/chaincode/objective"
	"github.com/owkin/orchestrator/utils"
)

// Static list of all contracts
var allContracts []contractapi.ContractInterface = []contractapi.ContractInterface{
	node.NewSmartContract(),
	objective.NewSmartContract(),
	datasample.NewSmartContract(),
	algo.NewSmartContract(),
	datamanager.NewSmartContract(),
	computetask.NewSmartContract(),
}

// Type definitions

type Provider interface {
	GetAllContracts() []contractapi.ContractInterface
}

type TransactionChecker interface {
	IsEvaluateMethod(method string) bool
}

type ContractCollection struct {
	allEvaluateTransactions []string
}

// Implementation

func NewContractCollection() *ContractCollection {
	allEvalTx := buildAllEvaluateTransactions()
	return &ContractCollection{
		allEvaluateTransactions: allEvalTx,
	}
}

func (c *ContractCollection) GetAllContracts() []contractapi.ContractInterface {
	return AllContracts
}

// IsEvaluateMethod returns true if the parameter 'method' matches one
// one the smart contract methods defined as "Evaluate-only" in the list of
// all smart-contracts
func (c *ContractCollection) IsEvaluateMethod(method string) bool {
	return utils.StringInSlice(c.allEvaluateTransactions, method)
}

func buildAllEvaluateTransactions() []string {
	res := make([]string, 0)
	allContracts := allContracts
	for _, c := range allContracts {
		contract := c.GetName()
		if eci, ok := c.(contractapi.EvaluationContractInterface); ok {
			for _, method := range eci.GetEvaluateTransactions() {
				res = append(res, fmt.Sprintf("%v:%v", contract, method))
			}
		}
	}
	return res
}
