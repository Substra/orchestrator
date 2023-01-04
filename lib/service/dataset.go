package service

import (
	"github.com/substra/orchestrator/lib/asset"
)

// DatasetAPI defines the methods to act on Datasets
type DatasetAPI interface {
	GetDataset(id string) (*asset.Dataset, error)
}

// DatasetServiceProvider defines an object able to provide an DatasetAPI instance
type DatasetServiceProvider interface {
	GetDatasetService() DatasetAPI
}

// DatasetDependencyProvider defines what the DatasetService needs to perform its duty
type DatasetDependencyProvider interface {
	DataManagerServiceProvider
	DataSampleServiceProvider
}

// DatasetService is the Dataset manipulation entry point
// it implements the API interface
type DatasetService struct {
	DatasetDependencyProvider
}

// NewDatasetService will create a new service with given persistence layer
func NewDatasetService(provider DatasetDependencyProvider) *DatasetService {
	return &DatasetService{provider}
}

// GetDataset retrieves a single Dataset by its ID
func (s *DatasetService) GetDataset(id string) (*asset.Dataset, error) {
	datamanager, err := s.GetDataManagerService().GetDataManager(id)
	if err != nil {
		return nil, err
	}

	trainDataSampleKeys, err := s.GetDataSampleService().GetDataSampleKeysByManager(id)
	if err != nil {
		return nil, err
	}

	testDataSampleKeys, err := s.GetDataSampleService().GetDataSampleKeysByManager(id)
	if err != nil {
		return nil, err
	}

	dataset := &asset.Dataset{
		DataManager:         datamanager,
		TrainDataSampleKeys: trainDataSampleKeys,
		TestDataSampleKeys:  testDataSampleKeys,
	}

	return dataset, nil
}
