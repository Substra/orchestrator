package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
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
func (a *NodeAdapter) RegisterNode(ctx context.Context, in *asset.RegisterNodeParam) (*asset.Node, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.node:RegisterNode"

	node := &asset.Node{}

	err = invocator.Call(method, in, node)

	return node, err
}

// GetAllNodes will return all known nodes
func (a *NodeAdapter) GetAllNodes(ctx context.Context, in *asset.GetAllNodesParam) (*asset.GetAllNodesResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.node:GetAllNodes"

	nodes := &asset.GetAllNodesResponse{}

	err = invocator.Call(method, in, nodes)

	return nodes, err
}
