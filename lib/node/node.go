package node

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
	"golang.org/x/net/context"
)

// Server is exported
type Server struct {
	dbFactory persistence.Factory
}

// RegisterNode is exported
func (s *Server) RegisterNode(ctx context.Context, n *Node) (*Node, error) {
	db, err := s.dbFactory(ctx)
	if err != nil {
		fmt.Errorf("Cannot derive DB from context: %v", err)
		return nil, err
	}
	log.Println(n)
	log.Printf("node: %s, %s, %s", n.GetId(), n.GetModelKey(), n.GetFoo())

	nodeBytes, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}

	err = db.PutState(n.Id, nodeBytes)
	return n, err
}
