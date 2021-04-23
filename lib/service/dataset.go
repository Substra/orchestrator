// Copyright 2021 Owkin Inc.
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
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/persistence"
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
	persistence.DatasetDBALProvider
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
	return s.GetDatasetDBAL().GetDataset(id)
}
