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
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/algo"
	"github.com/owkin/orchestrator/chaincode/computeplan"
	"github.com/owkin/orchestrator/chaincode/computetask"
	"github.com/owkin/orchestrator/chaincode/datamanager"
	"github.com/owkin/orchestrator/chaincode/datasample"
	"github.com/owkin/orchestrator/chaincode/dataset"
	"github.com/owkin/orchestrator/chaincode/event"
	"github.com/owkin/orchestrator/chaincode/model"
	"github.com/owkin/orchestrator/chaincode/node"
	"github.com/owkin/orchestrator/chaincode/objective"
	"github.com/owkin/orchestrator/chaincode/performance"
)

// AllContracts is the list referencing all smartcontracts supported by the chaincode
var AllContracts []contractapi.ContractInterface = []contractapi.ContractInterface{
	node.NewSmartContract(),
	objective.NewSmartContract(),
	datasample.NewSmartContract(),
	algo.NewSmartContract(),
	datamanager.NewSmartContract(),
	dataset.NewSmartContract(),
	computetask.NewSmartContract(),
	model.NewSmartContract(),
	computeplan.NewSmartContract(),
	performance.NewSmartContract(),
	event.NewSmartContract(),
}
