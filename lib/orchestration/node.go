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
	RegisterNode(*assets.Node) error
	GetNodes() ([]*assets.Node, error)
}

// NodeService is the node manipulation entry point
// it implements NodeAPI
type NodeService struct {
	db persistence.Database
}

// NewNodeService will create a new service with given persistence layer
func NewNodeService(db persistence.Database) *NodeService {
	return &NodeService{db: db}
}

// RegisterNode persist a node
func (s *NodeService) RegisterNode(n *assets.Node) error {
	nodeBytes, err := json.Marshal(n)
	if err != nil {
		return err
	}

	return s.db.PutState(nodeResource, n.GetId(), nodeBytes)
}

// GetNodes list all known nodes
func (s *NodeService) GetNodes() ([]*assets.Node, error) {
	b, err := s.db.GetAll(nodeResource)
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
