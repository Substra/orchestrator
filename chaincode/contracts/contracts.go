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
	"github.com/owkin/orchestrator/chaincode/failurereport"
	"github.com/owkin/orchestrator/chaincode/info"
	"github.com/owkin/orchestrator/chaincode/metric"
	"github.com/owkin/orchestrator/chaincode/model"
	"github.com/owkin/orchestrator/chaincode/node"
	"github.com/owkin/orchestrator/chaincode/performance"
)

// AllContracts is the list referencing all smartcontracts supported by the chaincode
var AllContracts []contractapi.ContractInterface = []contractapi.ContractInterface{
	node.NewSmartContract(),
	metric.NewSmartContract(),
	datasample.NewSmartContract(),
	algo.NewSmartContract(),
	datamanager.NewSmartContract(),
	dataset.NewSmartContract(),
	computetask.NewSmartContract(),
	model.NewSmartContract(),
	computeplan.NewSmartContract(),
	performance.NewSmartContract(),
	event.NewSmartContract(),
	info.NewSmartContract(),
	failurereport.NewSmartContract(),
}
