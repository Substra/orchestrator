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

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// NodeDBAL defines the database abstraction layer to manipulate nodes
type NodeDBAL interface {
	// AddNode stores a new node.
	AddNode(node *asset.Node) error
	// NodeExists returns whether a node with the given ID is already in store
	NodeExists(id string) (bool, error)
	// GetAllNodes returns all known nodes
	GetAllNodes() ([]*asset.Node, error)
	// GetNode returns a Node by its ID
	GetNode(id string) (*asset.Node, error)
}

// ObjectiveDBAL is the database abstraction layer for Objectives
type ObjectiveDBAL interface {
	AddObjective(obj *asset.Objective) error
	GetObjective(key string) (*asset.Objective, error)
	QueryObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error)
	ObjectiveExists(key string) (bool, error)
}

// DataSampleDBAL is the database abstraction layer for DataSamples
type DataSampleDBAL interface {
	AddDataSample(dataSample *asset.DataSample) error
	UpdateDataSample(dataSample *asset.DataSample) error
	GetDataSample(key string) (*asset.DataSample, error)
	QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error)
	DataSampleExists(key string) (bool, error)
}

// AlgoDBAL is the database abstraction layer for Algos
type AlgoDBAL interface {
	AddAlgo(obj *asset.Algo) error
	GetAlgo(key string) (*asset.Algo, error)
	QueryAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error)
	AlgoExists(key string) (bool, error)
}

// DataManagerDBAL is the database abstraction layer for DataManagers
type DataManagerDBAL interface {
	AddDataManager(datamanager *asset.DataManager) error
	UpdateDataManager(datamanager *asset.DataManager) error
	GetDataManager(key string) (*asset.DataManager, error)
	QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error)
	DataManagerExists(key string) (bool, error)
}

// DatasetDBAL is the database abstraction layer for Datasets
type DatasetDBAL interface {
	GetDataset(id string) (*asset.Dataset, error)
}

// NodeDBALProvider representes an object capable of providing a NodeDBAL
type NodeDBALProvider interface {
	GetNodeDBAL() NodeDBAL
}

// ObjectiveDBALProvider represents an object capable of providing an ObjectiveDBAL
type ObjectiveDBALProvider interface {
	GetObjectiveDBAL() ObjectiveDBAL
}

// DataSampleDBALProvider represents an object capable of providing a DataSampleDBAL
type DataSampleDBALProvider interface {
	GetDataSampleDBAL() DataSampleDBAL
}

// AlgoDBALProvider represents an object capable of providing an AlgoDBAL
type AlgoDBALProvider interface {
	GetAlgoDBAL() AlgoDBAL
}

// DataManagerDBALProvider represents an object capable of providing a DataManagerDBAL
type DataManagerDBALProvider interface {
	GetDataManagerDBAL() DataManagerDBAL
}

// DatasetDBALProvider represents an object capable of providing a DatasetDBAL
type DatasetDBALProvider interface {
	GetDatasetDBAL() DatasetDBAL
}

// DBAL stands for Database Abstraction Layer, it exposes methods to interact with asset storage.
type DBAL interface {
	NodeDBAL
	ObjectiveDBAL
	DataSampleDBAL
	AlgoDBAL
	DataManagerDBAL
	DatasetDBAL
	ComputeTaskDBAL
	ModelDBAL
	ComputePlanDBAL
	PerformanceDBAL
	EventDBAL
}
