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

// Package persistence holds everything related to data persistence.
// Each asset has its own database abstraction layer (DBAL).
// Note that one cannot read its own writes: ie AddObjective then GetObjective won't work.
// Each request is a transaction which is only commited once a successful response is returned.
package persistence

import "github.com/owkin/orchestrator/lib/assets"

// NodeDBAL defines the database abstraction layer to manipulate nodes
type NodeDBAL interface {
	// AddNode stores a new node.
	AddNode(node *assets.Node) error
	// NodeExists returns whether a node with the given ID is already in store
	NodeExists(id string) (bool, error)
	// GetNodes returns all known nodes
	GetNodes() ([]*assets.Node, error)
}

// ObjectiveDBAL is the database abstraction layer for Objectives
type ObjectiveDBAL interface {
	AddObjective(obj *assets.Objective) error
	GetObjective(id string) (*assets.Objective, error)
	GetObjectives() ([]*assets.Objective, error) // TODO: pagination
}

// NodeDBALProvider representes an object capable of providing a NodeDBAL
type NodeDBALProvider interface {
	GetNodeDBAL() NodeDBAL
}

// ObjectiveDBALProvider represents an object capable of providing an ObjectiveDBAL
type ObjectiveDBALProvider interface {
	GetObjectiveDBAL() ObjectiveDBAL
}

// DBAL stands for Database Abstraction Layer, it exposes methods to interact with asset storage.
type DBAL interface {
	NodeDBAL
	ObjectiveDBAL
}
