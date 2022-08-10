package contracts

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substra/orchestrator/utils"
)

// Type definitions

type Provider interface {
	GetAllContracts() []contractapi.ContractInterface
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
// of the smart contract methods defined as "Evaluate-only" in the list of
// all smart-contracts
func (c *ContractCollection) IsEvaluateMethod(method string) bool {
	return utils.SliceContains(c.allEvaluateTransactions, method)
}

func buildAllEvaluateTransactions() []string {
	res := make([]string, 0)
	for _, c := range AllContracts {
		contract := c.GetName()
		if eci, ok := c.(contractapi.EvaluationContractInterface); ok {
			for _, method := range eci.GetEvaluateTransactions() {
				res = append(res, fmt.Sprintf("%v:%v", contract, method))
			}
		}
	}
	return res
}
