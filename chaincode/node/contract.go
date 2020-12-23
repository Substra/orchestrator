package node

import (
	"context"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	nodeService "github.com/substrafoundation/substra-orchestrator/lib/assets/node"
)

// SmartContract manages nodes
type SmartContract struct {
	contractapi.Contract
	server *nodeService.Server
}

func NewNodeContract(server *nodeService.Server) *SmartContract {
	return &SmartContract{
		server: server,
	}
}

// RegisterNode creates a new node in world state
func (s *SmartContract) RegisterNode(ctx contractapi.TransactionContextInterface, id string) error {
	node := nodeService.Node{Id: id}

	_, err := s.server.RegisterNode(context.Background(), &node)

	return err
}
