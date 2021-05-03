// Copyright 2021 Owkin Inc.
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

package distributed

import (
	"github.com/owkin/orchestrator/lib/asset"
	"golang.org/x/net/context"
)

// NodeAdapter is a grpc server exposing the same node interface,
// but relies on a remote chaincode to actually manage the asset.
type NodeAdapter struct {
	asset.UnimplementedNodeServiceServer
}

// NewNodeAdapter creates a Server
func NewNodeAdapter() *NodeAdapter {
	return &NodeAdapter{}
}

// RegisterNode will add a new node to the network
func (a *NodeAdapter) RegisterNode(ctx context.Context, in *asset.NodeRegistrationParam) (*asset.Node, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.node:RegisterNode"

	node := &asset.Node{}

	err = invocator.Call(method, in, node)

	return node, err
}

// QueryNodes will return all known nodes
func (a *NodeAdapter) QueryNodes(ctx context.Context, in *asset.NodeQueryParam) (*asset.NodeQueryResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.node:QueryNodes"

	nodes := &asset.NodeQueryResponse{}

	err = invocator.Call(method, in, nodes)

	return nodes, err
}
