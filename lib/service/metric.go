package service

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/owkin/orchestrator/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MetricAPI defines the methods to act on Metrics
type MetricAPI interface {
	RegisterMetric(metric *asset.NewMetric, owner string) (*asset.Metric, error)
	GetMetric(string) (*asset.Metric, error)
	QueryMetrics(p *common.Pagination) ([]*asset.Metric, common.PaginationToken, error)
	CanDownload(key string, requester string) (bool, error)
}

// MetricServiceProvider defines an object able to provide an MetricAPI instance
type MetricServiceProvider interface {
	GetMetricService() MetricAPI
}

// MetricDependencyProvider defines what the MetricService needs to perform its duty
type MetricDependencyProvider interface {
	LoggerProvider
	persistence.MetricDBALProvider
	EventServiceProvider
	PermissionServiceProvider
	TimeServiceProvider
}

// MetricService is the metric manipulation entry point
// it implements the API interface
type MetricService struct {
	MetricDependencyProvider
}

// NewMetricService will create a new service with given persistence layer
func NewMetricService(provider MetricDependencyProvider) *MetricService {
	return &MetricService{provider}
}

// RegisterMetric persist an metric
func (s *MetricService) RegisterMetric(o *asset.NewMetric, owner string) (*asset.Metric, error) {
	s.GetLogger().WithField("owner", owner).WithField("newObj", o).Debug("Registering metric")
	err := o.Validate()
	if err != nil {
		return nil, orcerrors.FromValidationError(asset.MetricKind, err)
	}

	metric := &asset.Metric{
		Key:          o.Key,
		Name:         o.Name,
		Description:  o.Description,
		Address:      o.Address,
		Metadata:     o.Metadata,
		Owner:        owner,
		CreationDate: timestamppb.New(s.GetTimeService().GetTransactionTime()),
	}

	metric.Permissions, err = s.GetPermissionService().CreatePermissions(owner, o.NewPermissions)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  o.Key,
		AssetKind: asset.AssetKind_ASSET_METRIC,
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	err = s.GetMetricDBAL().AddMetric(metric)
	if err != nil {
		return nil, err
	}

	return metric, nil
}

// GetMetric retrieves an metric by its key
func (s *MetricService) GetMetric(key string) (*asset.Metric, error) {
	return s.GetMetricDBAL().GetMetric(key)
}

// QueryMetrics returns all stored metrics
func (s *MetricService) QueryMetrics(p *common.Pagination) ([]*asset.Metric, common.PaginationToken, error) {
	return s.GetMetricDBAL().QueryMetrics(p)
}

// CanDownload checks if the requester can download the metric corresponding to the provided key
func (s *MetricService) CanDownload(key string, requester string) (bool, error) {
	obj, err := s.GetMetric(key)

	if err != nil {
		return false, err
	}

	if obj.Permissions.Download.Public || utils.StringInSlice(obj.Permissions.Download.AuthorizedIds, requester) {
		return true, nil
	}

	return false, nil
}
