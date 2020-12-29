// Copyright 2020 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
