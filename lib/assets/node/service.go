package node

import (
	"encoding/json"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

const resource = "nodes"

// API defines the methods to act on Nodes
type API interface {
	RegisterNode(*Node) error
	GetNodes() ([]*Node, error)
}

// Service is the node manipulation entry point
// it implements the API
type Service struct {
	db persistence.Database
}

// NewService will create a new service with given persistence layer
func NewService(db persistence.Database) *Service {
	return &Service{db: db}
}

// RegisterNode persist a node
func (s *Service) RegisterNode(n *Node) error {
	nodeBytes, err := json.Marshal(n)
	if err != nil {
		return err
	}

	s.db.PutState(resource, n.GetId(), nodeBytes)

	return nil
}

// GetNodes list all known nodes
func (s *Service) GetNodes() ([]*Node, error) {
	b, err := s.db.GetAll(resource)
	if err != nil {
		return nil, err
	}

	var nodes []*Node

	for _, nodeBytes := range b {
		n := Node{}
		err = json.Unmarshal(nodeBytes, &n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	return nodes, nil
}
