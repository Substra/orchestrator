package service

import (
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/persistence"
)

// DataSampleAPI defines the methods to act on DataSamples
type DataSampleAPI interface {
	RegisterDataSamples(datasamples []*asset.NewDataSample, owner string) ([]*asset.DataSample, error)
	UpdateDataSamples(datasample *asset.UpdateDataSamplesParam, owner string) error
	QueryDataSamples(p *common.Pagination, filter *asset.DataSampleQueryFilter) ([]*asset.DataSample, common.PaginationToken, error)
	CheckSameManager(managerKey string, sampleKeys []string) error
	IsTestOnly(sampleKeys []string) (bool, error)
	ContainsTestSample(sampleKeys []string) (bool, error)
	GetDataSampleKeysByManager(managerKey string, testOnly bool) ([]string, error)
	GetDataSample(string) (*asset.DataSample, error)
}

// DataSampleServiceProvider defines an object able to provide a DataSampleAPI instance
type DataSampleServiceProvider interface {
	GetDataSampleService() DataSampleAPI
}

// DataSampleDependencyProvider defines what the DataSampleService needs to perform its duty
type DataSampleDependencyProvider interface {
	LoggerProvider
	persistence.DataSampleDBALProvider
	DataManagerServiceProvider
	EventServiceProvider
	TimeServiceProvider
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

func (s *DataSampleService) RegisterDataSamples(samples []*asset.NewDataSample, owner string) ([]*asset.DataSample, error) {
	s.GetLogger().WithField("owner", owner).WithField("nbSamples", len(samples)).Debug("Registering data samples")

	registeredSamples := []*asset.DataSample{}
	events := []*asset.Event{}

	for _, newSample := range samples {
		sample, err := s.createDataSample(newSample, owner)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	err = s.GetDataSampleDBAL().AddDataSamples(registeredSamples...)
	if err != nil {
		return nil, err
	}

	return registeredSamples, nil
}

// createDataSample persist one datasample
func (s *DataSampleService) createDataSample(sample *asset.NewDataSample, owner string) (*asset.DataSample, error) {
	s.GetLogger().WithField("owner", owner).WithField("newDataSample", sample).Debug("Registering data sample")
	err := sample.Validate()
	if err != nil {
		return nil, orcerrors.FromValidationError(asset.DataSampleKind, err)
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
		return nil, orcerrors.NewConflict(asset.DataSampleKind, sample.Key)
	}
	datasample := &asset.DataSample{
		Key:             sample.Key,
		DataManagerKeys: sample.GetDataManagerKeys(),
		TestOnly:        sample.GetTestOnly(),
		Owner:           owner,
		Checksum:        sample.Checksum,
		CreationDate:    timestamppb.New(s.GetTimeService().GetTransactionTime()),
	}

	return datasample, nil
}

// UpdateDataSamples update or add one or multiple datasamples
func (s *DataSampleService) UpdateDataSamples(d *asset.UpdateDataSamplesParam, owner string) error {
	s.GetLogger().WithField("owner", owner).WithField("dataSampleUpdate", d).Debug("Updating data sample")
	err := d.Validate()
	if err != nil {
		return orcerrors.FromValidationError(asset.DataSampleKind, err)
	}

	err = s.GetDataManagerService().CheckOwner(d.GetDataManagerKeys(), owner)
	if err != nil {
		return err
	}

	for _, dataSampleKey := range d.GetKeys() {
		datasample, err := s.GetDataSampleDBAL().GetDataSample(dataSampleKey)
		if err != nil {
			return orcerrors.NewNotFound(asset.DataSampleKind, dataSampleKey)
		}

		if datasample.GetOwner() != owner {
			return orcerrors.NewPermissionDenied("requester does not own the datasample")
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
func (s *DataSampleService) QueryDataSamples(p *common.Pagination, filter *asset.DataSampleQueryFilter) ([]*asset.DataSample, common.PaginationToken, error) {
	return s.GetDataSampleDBAL().QueryDataSamples(p, filter)
}

// CheckSameManager validates that samples all have in common the given manager.
func (s *DataSampleService) CheckSameManager(managerKey string, sampleKeys []string) error {
	for _, sampleKey := range sampleKeys {
		dataSample, err := s.GetDataSampleDBAL().GetDataSample(sampleKey)
		if err != nil {
			return err
		}
		if !utils.StringInSlice(dataSample.DataManagerKeys, managerKey) {
			return orcerrors.NewInvalidAsset("datasamples do not share a common manager")
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

func (s *DataSampleService) GetDataSampleKeysByManager(managerKey string, testOnly bool) ([]string, error) {
	return s.GetDataSampleDBAL().GetDataSampleKeysByManager(managerKey, testOnly)
}

// GetDataSample retrieves an datasample by its key
func (s *DataSampleService) GetDataSample(key string) (*asset.DataSample, error) {
	return s.GetDataSampleDBAL().GetDataSample(key)
}
