// Package persistence holds everything related to data persistence.
// Each asset has its own database abstraction layer (DBAL).
// Each request is a transaction which is only committed once a successful response is returned.
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

// MetricDBAL is the database abstraction layer for Metrics
type MetricDBAL interface {
	AddMetric(obj *asset.Metric) error
	GetMetric(key string) (*asset.Metric, error)
	QueryMetrics(p *common.Pagination) ([]*asset.Metric, common.PaginationToken, error)
	MetricExists(key string) (bool, error)
}

// DataSampleDBAL is the database abstraction layer for DataSamples
type DataSampleDBAL interface {
	AddDataSamples(dataSample ...*asset.DataSample) error
	UpdateDataSample(dataSample *asset.DataSample) error
	GetDataSample(key string) (*asset.DataSample, error)
	QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error)
	DataSampleExists(key string) (bool, error)
	GetDataSampleKeysByManager(managerKey string, testOnly bool) ([]string, error)
}

// AlgoDBAL is the database abstraction layer for Algos
type AlgoDBAL interface {
	AddAlgo(obj *asset.Algo) error
	GetAlgo(key string) (*asset.Algo, error)
	QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error)
	AlgoExists(key string) (bool, error)
}

// DataManagerDBAL is the database abstraction layer for DataManagers
type DataManagerDBAL interface {
	AddDataManager(datamanager *asset.DataManager) error
	GetDataManager(key string) (*asset.DataManager, error)
	QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error)
	DataManagerExists(key string) (bool, error)
}

// NodeDBALProvider represents an object capable of providing a NodeDBAL
type NodeDBALProvider interface {
	GetNodeDBAL() NodeDBAL
}

// MetricDBALProvider represents an object capable of providing an MetricDBAL
type MetricDBALProvider interface {
	GetMetricDBAL() MetricDBAL
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

// DBAL stands for Database Abstraction Layer, it exposes methods to interact with asset storage.
type DBAL interface {
	NodeDBAL
	MetricDBAL
	DataSampleDBAL
	AlgoDBAL
	DataManagerDBAL
	ComputeTaskDBAL
	ModelDBAL
	ComputePlanDBAL
	PerformanceDBAL
	EventDBAL
	FailureReportDBAL
}

// DBALProvider exposes all available DBAL.
type DBALProvider interface {
	NodeDBALProvider
	MetricDBALProvider
	DataSampleDBALProvider
	AlgoDBALProvider
	DataManagerDBALProvider
	ComputeTaskDBALProvider
	ModelDBALProvider
	ComputePlanDBALProvider
	PerformanceDBALProvider
	EventDBALProvider
	FailureReportDBALProvider
}
