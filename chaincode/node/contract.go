package node

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/substrafoundation/substra-orchestrator/chaincode/ledger"
	nodeAsset "github.com/substrafoundation/substra-orchestrator/lib/assets/node"
)

func getServiceFromContext(ctx contractapi.TransactionContextInterface) (nodeAsset.API, error) {
	db, err := ledger.GetLedgerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return nodeAsset.NewService(db), nil
}

// SmartContract manages nodes
type SmartContract struct {
	contractapi.Contract
	serviceFactory func(contractapi.TransactionContextInterface) (nodeAsset.API, error)
}

func NewSmartContract() *SmartContract {
	return &SmartContract{
		serviceFactory: getServiceFromContext,
	}
}

// NodeResponse represents a Node
type NodeResponse struct {
	Id string `json:"id"`
}

func responseFromAsset(n *nodeAsset.Node) *NodeResponse {
	return &NodeResponse{
		Id: n.Id,
	}
}

// RegisterNode creates a new node in world state
func (s *SmartContract) RegisterNode(ctx contractapi.TransactionContextInterface, id string) error {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return err
	}

	node := nodeAsset.Node{Id: id}

	err = service.RegisterNode(&node)
	return err
}

func (s *SmartContract) QueryNodes(ctx contractapi.TransactionContextInterface) ([]*NodeResponse, error) {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return nil, err
	}

	var r []*NodeResponse

	nodes, err := service.GetNodes()

	for _, n := range nodes {
		r = append(r, responseFromAsset(n))
	}

	return r, err
}
