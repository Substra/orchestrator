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

package standalone

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
)

// NodeServer is the gRPC server exposing node actions
type NodeServer struct {
	asset.UnimplementedNodeServiceServer
}

// NewNodeServer creates a Server
func NewNodeServer() *NodeServer {
	return &NodeServer{}
}

// RegisterNode will add a new node to the network
func (s *NodeServer) RegisterNode(ctx context.Context, in *asset.RegisterNodeParam) (*asset.Node, error) {
	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	node, err := services.GetNodeService().RegisterNode(mspid)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// GetAllNodes will return all known nodes
func (s *NodeServer) GetAllNodes(ctx context.Context, in *asset.GetAllNodesParam) (*asset.GetAllNodesResponse, error) {
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	nodes, err := services.GetNodeService().GetAllNodes()

	return &asset.GetAllNodesResponse{
		Nodes: nodes,
	}, err
}
