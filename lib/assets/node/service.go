package node

import (
	"encoding/json"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

// API defines the methods to act on Nodes
type API interface {
	RegisterNode(*Node) error
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

func (s *Service) RegisterNode(n *Node) error {
	nodeBytes, err := json.Marshal(n)
	if err != nil {
		return err
	}

	s.db.PutState(n.GetId(), nodeBytes)

	return nil
}
