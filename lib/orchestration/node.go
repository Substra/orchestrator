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

package orchestration

import (
	"encoding/json"

	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/persistence"
)

const nodeResource = "nodes"

// NodeAPI defines the methods to act on Nodes
type NodeAPI interface {
	RegisterNode(id string) (*assets.Node, error)
	GetNodes() ([]*assets.Node, error)
}

// NodeServiceProvider defines an object able to provide a NodeAPI instance
type NodeServiceProvider interface {
	GetNodeService() NodeAPI
}

// NodeDependencyProvider defines what the NodeService needs to perform its duty
type NodeDependencyProvider interface {
	persistence.DatabaseProvider
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
func (s *NodeService) RegisterNode(id string) (*assets.Node, error) {
	node := &assets.Node{Id: id}
	nodeBytes, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}

	err = s.GetDatabase().PutState(nodeResource, node.GetId(), nodeBytes)
	return node, err
}

// GetNodes list all known nodes
func (s *NodeService) GetNodes() ([]*assets.Node, error) {
	b, err := s.GetDatabase().GetAll(nodeResource)
	if err != nil {
		return nil, err
	}

	var nodes []*assets.Node

	for _, nodeBytes := range b {
		n := assets.Node{}
		err = json.Unmarshal(nodeBytes, &n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	return nodes, nil
}
