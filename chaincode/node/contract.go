package node

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substrafoundation/substra-orchestrator/chaincode/ledger"
	nodeAsset "github.com/substrafoundation/substra-orchestrator/lib/assets/node"
)

// SmartContract manages nodes
type SmartContract struct {
	contractapi.Contract
}

// RegisterNode creates a new node in world state
func (s *SmartContract) RegisterNode(ctx contractapi.TransactionContextInterface, id string) error {
	db, err := ledger.GetLedgerFromContext(ctx)
	if err != nil {
		return err
	}

	service := nodeAsset.NewService(db)
	node := nodeAsset.Node{Id: id}

	err = service.RegisterNode(&node)
	return err
}
