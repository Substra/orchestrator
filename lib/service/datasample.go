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

	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/persistence"
)

// DataSampleAPI defines the methods to act on DataSamples
type DataSampleAPI interface {
	RegisterDataSamples(datasamples []*asset.NewDataSample, owner string) error
	UpdateDataSamples(datasample *asset.UpdateDataSamplesParam, owner string) error
	QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error)
	CheckSameManager(managerKey string, sampleKeys []string) error
	IsTestOnly(sampleKeys []string) (bool, error)
	ContainsTestSample(sampleKeys []string) (bool, error)
}

// DataSampleServiceProvider defines an object able to provide a DataSampleAPI instance
type DataSampleServiceProvider interface {
	GetDataSampleService() DataSampleAPI
}

// DataSampleDependencyProvider defines what the DataSampleService needs to perform its duty
type DataSampleDependencyProvider interface {
	persistence.DataSampleDBALProvider
	DataManagerServiceProvider
	EventServiceProvider
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

func (s *DataSampleService) RegisterDataSamples(samples []*asset.NewDataSample, owner string) error {
	log.WithField("owner", owner).WithField("nbSamples", len(samples)).Debug("Registering data samples")

	registeredSamples := []*asset.DataSample{}
	events := []*asset.Event{}

	for _, newSample := range samples {
		sample, err := s.createDataSample(newSample, owner)
		if err != nil {
			return err
		}
		registeredSamples = append(registeredSamples, sample)

		event := &asset.Event{
			EventKind: asset.EventKind_EVENT_ASSET_CREATED,
			AssetKey:  sample.Key,
			AssetKind: asset.AssetKind_ASSET_DATA_SAMPLE,
		}
		events = append(events, event)

	}
	err := s.GetEventService().RegisterEvents(events...)
	if err != nil {
		return err
	}

	err = s.GetDataSampleDBAL().AddDataSamples(registeredSamples...)
	if err != nil {
		return err
	}

	return nil
}

// registerDataSample persist one datasamples
func (s *DataSampleService) createDataSample(sample *asset.NewDataSample, owner string) (*asset.DataSample, error) {
	log.WithField("owner", owner).WithField("newDataSample", sample).Debug("Registering data sample")
	err := sample.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", orcerrors.ErrInvalidAsset, err.Error())
	}

	err = s.GetDataManagerService().CheckOwner(sample.GetDataManagerKeys(), owner)
	if err != nil {
		return nil, err
	}

	exists, err := s.GetDataSampleDBAL().DataSampleExists(sample.Key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("datasample whith the same key already exist: %w key: %s", orcerrors.ErrConflict, sample.Key)
	}
	datasample := &asset.DataSample{
		Key:             sample.Key,
		DataManagerKeys: sample.GetDataManagerKeys(),
		TestOnly:        sample.GetTestOnly(),
		Owner:           owner,
		Checksum:        sample.Checksum,
	}

	return datasample, nil
}

// UpdateDataSamples update or add one or multiple datasamples
func (s *DataSampleService) UpdateDataSamples(d *asset.UpdateDataSamplesParam, owner string) error {
	log.WithField("owner", owner).WithField("dataSampleUpdate", d).Debug("Updating data sample")
	err := d.Validate()
	if err != nil {
		return fmt.Errorf("%w: %s", orcerrors.ErrInvalidAsset, err.Error())
	}

	err = s.GetDataManagerService().CheckOwner(d.GetDataManagerKeys(), owner)
	if err != nil {
		return err
	}

	for _, dataSampleKey := range d.GetKeys() {
		datasample, err := s.GetDataSampleDBAL().GetDataSample(dataSampleKey)
		if err != nil {
			return fmt.Errorf("datasample not found: %w key: %s ", orcerrors.ErrNotFound, dataSampleKey)
		}

		if datasample.GetOwner() != owner {
			return fmt.Errorf("requester does not own the datasample: %w", orcerrors.ErrPermissionDenied)
		}

		datasample.DataManagerKeys = utils.Combine(datasample.GetDataManagerKeys(), d.GetDataManagerKeys())

		event := &asset.Event{
			EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
			AssetKey:  dataSampleKey,
			AssetKind: asset.AssetKind_ASSET_DATA_SAMPLE,
		}
		err = s.GetEventService().RegisterEvents(event)
		if err != nil {
			return err
		}

		err = s.GetDataSampleDBAL().UpdateDataSample(datasample)
		if err != nil {
			return err
		}
	}

	return nil
}

// QueryDataSamples returns all stored datasamples
func (s *DataSampleService) QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	return s.GetDataSampleDBAL().QueryDataSamples(p)
}

// CheckSameManager validates that samples all have in common the given manager.
func (s *DataSampleService) CheckSameManager(managerKey string, sampleKeys []string) error {
	for _, sampleKey := range sampleKeys {
		dataSample, err := s.GetDataSampleDBAL().GetDataSample(sampleKey)
		if err != nil {
			return err
		}
		if !utils.StringInSlice(dataSample.DataManagerKeys, managerKey) {
			return fmt.Errorf("datasamples do not share a common manager: %w", orcerrors.ErrInvalidAsset)
		}
	}
	return nil
}

// IsTestOnly returns if givens samples are for sanctuarized test data
func (s *DataSampleService) IsTestOnly(sampleKeys []string) (bool, error) {
	testOnly := true
	for _, sampleKey := range sampleKeys {
		dataSample, err := s.GetDataSampleDBAL().GetDataSample(sampleKey)
		if err != nil {
			return false, err
		}
		testOnly = testOnly && dataSample.TestOnly
	}
	return testOnly, nil
}

// ContainsTestSample returns true if there is at least a test sample in the list
func (s *DataSampleService) ContainsTestSample(sampleKeys []string) (bool, error) {
	hasTest := false
	for _, sampleKey := range sampleKeys {
		dataSample, err := s.GetDataSampleDBAL().GetDataSample(sampleKey)
		if err != nil {
			return false, err
		}
		hasTest = hasTest || dataSample.TestOnly
	}
	return hasTest, nil
}
