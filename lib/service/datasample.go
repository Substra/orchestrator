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
	"fmt"

	orchestrationErrors "github.com/owkin/orchestrator/lib/errors"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/persistence"
)

// DataSampleAPI defines the methods to act on DataSamples
type DataSampleAPI interface {
	RegisterDataSample(datasample *asset.NewDataSample, owner string) error
	UpdateDataSample(datasample *asset.DataSampleUpdateParam, owner string) error
	GetDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error)
}

// DataSampleServiceProvider defines an object able to provide a DataSampleAPI instance
type DataSampleServiceProvider interface {
	GetDataSampleService() DataSampleAPI
}

// DataSampleDependencyProvider defines what the DataSampleService needs to perform its duty
type DataSampleDependencyProvider interface {
	persistence.DataSampleDBALProvider
	DataManagerServiceProvider
}

// DataSampleService is the data samples manipulation entry point
// it implements the API interface
type DataSampleService struct {
	DataSampleDependencyProvider
}

// NewDataSampleService will create a new service with given dependency provider
func NewDataSampleService(provider DataSampleDependencyProvider) *DataSampleService {
	return &DataSampleService{provider}
}

// RegisterDataSample persist one or multiple datasamples
func (s *DataSampleService) RegisterDataSample(d *asset.NewDataSample, owner string) error {
	log.WithField("owner", owner).WithField("newDataSample", d).Debug("Registering data sample")
	err := d.Validate()
	if err != nil {
		return fmt.Errorf("%w: %s", orchestrationErrors.ErrInvalidAsset, err.Error())
	}

	err = s.GetDataManagerService().CheckOwner(d.GetDataManagerKeys(), owner)
	if err != nil {
		return err
	}

	for _, dataSampleKey := range d.GetKeys() {
		datasample := &asset.DataSample{
			Key:             dataSampleKey,
			DataManagerKeys: d.GetDataManagerKeys(),
			TestOnly:        d.GetTestOnly(),
			Owner:           owner,
		}

		err = s.GetDataSampleDBAL().AddDataSample(datasample)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateDataSample update or add one or multiple datasamples
func (s *DataSampleService) UpdateDataSample(d *asset.DataSampleUpdateParam, owner string) error {
	log.WithField("owner", owner).WithField("dataSampleUpdate", d).Debug("Updating data sample")
	err := d.Validate()
	if err != nil {
		return fmt.Errorf("%w: %s", orchestrationErrors.ErrInvalidAsset, err.Error())
	}

	err = s.GetDataManagerService().CheckOwner(d.GetDataManagerKeys(), owner)
	if err != nil {
		return err
	}

	for _, dataSampleKey := range d.GetKeys() {
		datasample, err := s.GetDataSampleDBAL().GetDataSample(dataSampleKey)
		if err != nil {
			return fmt.Errorf("datasample not found: %w key: %s ", orchestrationErrors.ErrNotFound, dataSampleKey)
		}

		if datasample.GetOwner() != owner {
			return fmt.Errorf("Requester does not own the datasample: %w", orchestrationErrors.ErrPermissionDenied)
		}

		datasample.DataManagerKeys = d.GetDataManagerKeys()

		err = s.GetDataSampleDBAL().UpdateDataSample(datasample)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetDataSamples returns all stored datasamples
func (s *DataSampleService) GetDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	return s.GetDataSampleDBAL().GetDataSamples(p)
}
