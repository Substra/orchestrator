package node

import (
	"log"

	"golang.org/x/net/context"
)

// Server is the gRPC server exposing node actions
type Server struct {
	UnimplementedNodeServiceServer
	nodeService *Service
}

// NewServer creates a Server
func NewServer(service *Service) *Server {
	return &Server{nodeService: service}
}

// RegisterNode will add a new node to the network
func (s *Server) RegisterNode(ctx context.Context, n *Node) (*Node, error) {
	log.Println(n)
	log.Printf("node: %s, %s, %s", n.GetId(), n.GetModelKey(), n.GetFoo())

	err := s.nodeService.RegisterNode(n)
	return n, err
}
