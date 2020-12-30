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

// Package node defines the Node asset and its business logic.
// A Node is an actor of the network.
package node

import (
	"encoding/json"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

const resource = "nodes"

// API defines the methods to act on Nodes
type API interface {
	RegisterNode(*Node) error
	GetNodes() ([]*Node, error)
}

// Service is the node manipulation entry point
// it implements the API
type Service struct {
	db persistence.Database
}

// NewService will create a new service with given persistence layer
func NewService(db persistence.Database) *Service {
	return &Service{db: db}
}

// RegisterNode persist a node
func (s *Service) RegisterNode(n *Node) error {
	nodeBytes, err := json.Marshal(n)
	if err != nil {
		return err
	}

	s.db.PutState(resource, n.GetId(), nodeBytes)

	return nil
}

// GetNodes list all known nodes
func (s *Service) GetNodes() ([]*Node, error) {
	b, err := s.db.GetAll(resource)
	if err != nil {
		return nil, err
	}

	var nodes []*Node

	for _, nodeBytes := range b {
		n := Node{}
		err = json.Unmarshal(nodeBytes, &n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	return nodes, nil
}
