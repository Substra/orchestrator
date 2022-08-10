package service

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ComputePlanAPI defines the methods to act on ComputePlans
type ComputePlanAPI interface {
	RegisterPlan(plan *asset.NewComputePlan, owner string) (*asset.ComputePlan, error)
	GetPlan(key string) (*asset.ComputePlan, error)
	QueryPlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error)
	ApplyPlanAction(key string, action asset.ComputePlanAction, requester string) error
	UpdatePlan(computePlan *asset.UpdateComputePlanParam, requester string) error
	canDeleteModels(key string) (bool, error)
	computePlanExists(key string) (bool, error)
}

// ComputePlanServiceProvider defines an object able to provide a ComputePlanAPI instance
type ComputePlanServiceProvider interface {
	GetComputePlanService() ComputePlanAPI
}

// ComputePlanDependencyProvider defines what the ComputePlanService needs to perform its duty
type ComputePlanDependencyProvider interface {
	LoggerProvider
	persistence.ComputePlanDBALProvider
	EventServiceProvider
	ComputeTaskServiceProvider
	TimeServiceProvider
}

// ComputePlanService is the compute plan manipulation entry point
type ComputePlanService struct {
	ComputePlanDependencyProvider
}

// NewComputePlanService creates a new service
func NewComputePlanService(provider ComputePlanDependencyProvider) *ComputePlanService {
	return &ComputePlanService{provider}
}

func (s *ComputePlanService) RegisterPlan(input *asset.NewComputePlan, owner string) (*asset.ComputePlan, error) {
	s.GetLogger().WithField("plan", input).WithField("owner", owner).Debug("Registering new compute plan")
	err := input.Validate()
	if err != nil {
		return nil, orcerrors.FromValidationError(asset.ComputePlanKind, err)
	}

	exist, err := s.GetComputePlanDBAL().ComputePlanExists(input.Key)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, orcerrors.NewConflict(asset.ComputePlanKind, input.Key)
	}

	plan := &asset.ComputePlan{
		Key:                      input.Key,
		Owner:                    owner,
		Tag:                      input.Tag,
		Name:                     input.Name,
		Metadata:                 input.Metadata,
		DeleteIntermediaryModels: input.DeleteIntermediaryModels,
		CreationDate:             timestamppb.New(s.GetTimeService().GetTransactionTime()),
		Status:                   asset.ComputePlanStatus_PLAN_STATUS_EMPTY,
	}

	err = s.GetComputePlanDBAL().AddComputePlan(plan)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  plan.Key,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		Asset:     &asset.Event_ComputePlan{ComputePlan: plan},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *ComputePlanService) ApplyPlanAction(key string, action asset.ComputePlanAction, requester string) error {
	plan, err := s.GetComputePlanDBAL().GetComputePlan(key)
	if err != nil {
		return err
	}
	if requester != plan.Owner {
		return orcerrors.NewPermissionDenied("only plan owner can act on it")
	}

	switch action {
	case asset.ComputePlanAction_PLAN_ACTION_CANCELED:
		return s.cancelPlan(plan)
	default:
		return orcerrors.NewUnimplemented("plan action unimplemented")
	}
}

func (s *ComputePlanService) GetPlan(key string) (*asset.ComputePlan, error) {
	return s.GetComputePlanDBAL().GetComputePlan(key)
}

func (s *ComputePlanService) QueryPlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error) {
	return s.GetComputePlanDBAL().QueryComputePlans(p, filter)
}

// UpdatePlan updates mutable fields of a compute plan. List of mutable fields : name.
func (s *ComputePlanService) UpdatePlan(a *asset.UpdateComputePlanParam, requester string) error {
	s.GetLogger().WithField("requester", requester).WithField("computePlanUpdate", a).Debug("Updating compute plan")
	err := a.Validate()
	if err != nil {
		return orcerrors.FromValidationError(asset.ComputePlanKind, err)
	}

	planKey := a.Key

	plan, err := s.GetComputePlanDBAL().GetComputePlan(planKey)
	if err != nil {
		return orcerrors.NewNotFound(asset.ComputePlanKind, planKey)
	}

	if requester != plan.Owner {
		return orcerrors.NewPermissionDenied("requester does not own the compute plan")
	}

	// Update compute plan name
	plan.Name = a.Name

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  planKey,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		Asset:     &asset.Event_ComputePlan{ComputePlan: plan},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return err
	}

	return s.GetComputePlanDBAL().UpdateComputePlan(plan)
}

func (s *ComputePlanService) cancelPlan(plan *asset.ComputePlan) error {

	if plan.CancelationDate != nil {
		return orcerrors.NewBadRequest("compute plan is already canceled")
	}

	txTimestamp := s.GetTimeService().GetTransactionTime()

	err := s.GetComputePlanDBAL().CancelComputePlan(plan, txTimestamp)
	if err != nil {
		return err
	}

	plan.CancelationDate = timestamppb.New(txTimestamp)
	plan.Status = asset.ComputePlanStatus_PLAN_STATUS_CANCELED

	event := &asset.Event{
		AssetKey:  plan.Key,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		Asset:     &asset.Event_ComputePlan{ComputePlan: plan},
	}

	return s.GetEventService().RegisterEvents(event)
}

// canDeleteModels returns true if the compute plan allows intermediary models deletion.
func (s *ComputePlanService) canDeleteModels(key string) (bool, error) {
	plan, err := s.GetComputePlanDBAL().GetRawComputePlan(key)
	if err != nil {
		return false, err
	}

	return plan.DeleteIntermediaryModels, nil
}

func (s *ComputePlanService) computePlanExists(key string) (bool, error) {
	return s.GetComputePlanDBAL().ComputePlanExists(key)
}
