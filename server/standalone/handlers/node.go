package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
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
	services, err := interceptors.ExtractProvider(ctx)
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
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	nodes, err := services.GetNodeService().GetAllNodes()

	return &asset.GetAllNodesResponse{
		Nodes: nodes,
	}, err
}
