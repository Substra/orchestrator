package node

import (
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
	err := s.nodeService.RegisterNode(n)
	return n, err
}

// QueryNodes will return all known nodes
func (s *Server) QueryNodes(ctx context.Context, in *NodeQueryParam) (*NodeQueryResponse, error) {
	nodes, err := s.nodeService.GetNodes()

	return &NodeQueryResponse{
		Nodes: nodes,
	}, err
}
