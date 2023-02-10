package contracts

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substra/orchestrator/chaincode/computeplan"
	"github.com/substra/orchestrator/chaincode/computetask"
	"github.com/substra/orchestrator/chaincode/datamanager"
	"github.com/substra/orchestrator/chaincode/datasample"
	"github.com/substra/orchestrator/chaincode/dataset"
	"github.com/substra/orchestrator/chaincode/event"
	"github.com/substra/orchestrator/chaincode/failurereport"
	"github.com/substra/orchestrator/chaincode/function"
	"github.com/substra/orchestrator/chaincode/info"
	"github.com/substra/orchestrator/chaincode/model"
	"github.com/substra/orchestrator/chaincode/organization"
	"github.com/substra/orchestrator/chaincode/performance"
)

// AllContracts is the list referencing all smartcontracts supported by the chaincode
var AllContracts []contractapi.ContractInterface = []contractapi.ContractInterface{
	organization.NewSmartContract(),
	datasample.NewSmartContract(),
	function.NewSmartContract(),
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
