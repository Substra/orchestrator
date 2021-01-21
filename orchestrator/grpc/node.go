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

package grpc

import (
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/orchestration"
	"golang.org/x/net/context"
)

// NodeServer is the gRPC server exposing node actions
type NodeServer struct {
	assets.UnimplementedNodeServiceServer
	nodeService orchestration.NodeAPI
}

// NewNodeServer creates a Server
func NewNodeServer(service orchestration.NodeAPI) *NodeServer {
	return &NodeServer{nodeService: service}
}

// RegisterNode will add a new node to the network
func (s *NodeServer) RegisterNode(ctx context.Context, n *assets.Node) (*assets.Node, error) {
	err := s.nodeService.RegisterNode(n)
	return n, err
}

// QueryNodes will return all known nodes
func (s *NodeServer) QueryNodes(ctx context.Context, in *assets.NodeQueryParam) (*assets.NodeQueryResponse, error) {
	nodes, err := s.nodeService.GetNodes()

	return &assets.NodeQueryResponse{
		Nodes: nodes,
	}, err
}
