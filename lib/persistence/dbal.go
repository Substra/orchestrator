// Package persistence holds everything related to data persistence.
// Each asset has its own database abstraction layer (DBAL).
// Each request is a transaction which is only committed once a successful response is returned.
package persistence

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
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
	GetDataSampleKeysByManager(managerKey string) ([]string, error)
}

// FunctionDBAL is the database abstraction layer for Functions
type FunctionDBAL interface {
	AddFunction(obj *asset.Function) error
	GetFunction(key string) (*asset.Function, error)
	QueryFunctions(p *common.Pagination, filter *asset.FunctionQueryFilter) ([]*asset.Function, common.PaginationToken, error)
	FunctionExists(key string) (bool, error)
	UpdateFunction(function *asset.Function) error
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

// FunctionDBALProvider represents an object capable of providing an FunctionDBAL
type FunctionDBALProvider interface {
	GetFunctionDBAL() FunctionDBAL
}

// DataManagerDBALProvider represents an object capable of providing a DataManagerDBAL
type DataManagerDBALProvider interface {
	GetDataManagerDBAL() DataManagerDBAL
}

// DBAL stands for Database Abstraction Layer, it exposes methods to interact with asset storage.
type DBAL interface {
	OrganizationDBAL
	DataSampleDBAL
	FunctionDBAL
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
	FunctionDBALProvider
	DataManagerDBALProvider
	ComputeTaskDBALProvider
	ModelDBALProvider
	ComputePlanDBALProvider
	PerformanceDBALProvider
	EventDBALProvider
	FailureReportDBALProvider
}
