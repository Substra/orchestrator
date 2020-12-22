package node

import (
	"log"

	"golang.org/x/net/context"
)

// Server is exported
type Server struct{}

// RegisterNode is exported
func (s *Server) RegisterNode(ctx context.Context, n *Node) (*Node, error) {
	log.Println(n)
	log.Printf("node: %s, %s, %s", n.GetId(), n.GetModelKey(), n.GetFoo())

	return n, nil
}
