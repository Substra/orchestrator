// Package persistence holds everything related to data persistence.
// Each asset has its own database abstraction layer (DBAL).
// Each request is a transaction which is only committed once a successful response is returned.
package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// OrganizationDBAL defines the database abstraction layer to manipulate organizations
type OrganizationDBAL interface {
	// AddOrganization stores a new organization.
	AddOrganization(organization *asset.Organization) error
	// OrganizationExists returns whether an organization with the given ID is already in store
	OrganizationExists(id string) (bool, error)
	// GetAllOrganizations returns all known organizations
	GetAllOrganizations() ([]*asset.Organization, error)
	// GetOrganization returns an Organization by its ID
	GetOrganization(id string) (*asset.Organization, error)
}

// DataSampleDBAL is the database abstraction layer for DataSamples
type DataSampleDBAL interface {
	AddDataSamples(dataSample ...*asset.DataSample) error
	UpdateDataSample(dataSample *asset.DataSample) error
	GetDataSample(key string) (*asset.DataSample, error)
	QueryDataSamples(p *common.Pagination, filter *asset.DataSampleQueryFilter) ([]*asset.DataSample, common.PaginationToken, error)
	DataSampleExists(key string) (bool, error)
	GetDataSampleKeysByManager(managerKey string, testOnly bool) ([]string, error)
}

// AlgoDBAL is the database abstraction layer for Algos
type AlgoDBAL interface {
	AddAlgo(obj *asset.Algo) error
	GetAlgo(key string) (*asset.Algo, error)
	QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error)
	AlgoExists(key string) (bool, error)
	UpdateAlgo(algo *asset.Algo) error
}

// DataManagerDBAL is the database abstraction layer for DataManagers
type DataManagerDBAL interface {
	AddDataManager(datamanager *asset.DataManager) error
	GetDataManager(key string) (*asset.DataManager, error)
	QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error)
	DataManagerExists(key string) (bool, error)
	UpdateDataManager(dm *asset.DataManager) error
}

// OrganizationDBALProvider represents an object capable of providing an OrganizationDBAL
type OrganizationDBALProvider interface {
	GetOrganizationDBAL() OrganizationDBAL
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
	OrganizationDBAL
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
	OrganizationDBALProvider
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
