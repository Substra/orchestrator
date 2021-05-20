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

package service

import (
	"fmt"

	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
)

// NodeAPI defines the methods to act on Nodes
type NodeAPI interface {
	RegisterNode(id string) (*asset.Node, error)
	GetAllNodes() ([]*asset.Node, error)
	GetNode(id string) (*asset.Node, error)
}

// NodeServiceProvider defines an object able to provide a NodeAPI instance
type NodeServiceProvider interface {
	GetNodeService() NodeAPI
}

// NodeDependencyProvider defines what the NodeService needs to perform its duty
type NodeDependencyProvider interface {
	persistence.NodeDBALProvider
	EventServiceProvider
}

// NodeService is the node manipulation entry point
// it implements NodeAPI
type NodeService struct {
	NodeDependencyProvider
}

// NewNodeService will create a new service with given persistence layer
func NewNodeService(provider NodeDependencyProvider) *NodeService {
	return &NodeService{provider}
}

// RegisterNode persist a node
func (s *NodeService) RegisterNode(id string) (*asset.Node, error) {
	node := &asset.Node{Id: id}

	exists, err := s.GetNodeDBAL().NodeExists(id)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("node %s already exists: %w", node.GetId(), orcerrors.ErrConflict)
	}

	err = s.GetNodeDBAL().AddNode(node)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  id,
		AssetKind: asset.AssetKind_ASSET_NODE,
	}
	err = s.GetEventService().RegisterEvent(event)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// GetAllNodes list all known nodes
func (s *NodeService) GetAllNodes() ([]*asset.Node, error) {
	return s.GetNodeDBAL().GetAllNodes()
}

// GetNode returns a Node by its ID
func (s *NodeService) GetNode(id string) (*asset.Node, error) {
	return s.GetNodeDBAL().GetNode(id)
}
