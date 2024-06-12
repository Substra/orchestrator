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
	failPlan(key string) error
	computePlanExists(key string) (bool, error)
	IsPlanRunning(key string) (bool, error)
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
	s.GetLogger().Debug().Interface("plan", input).Str("owner", owner).Msg("Registering new compute plan")
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
		Key:          input.Key,
		Owner:        owner,
		Tag:          input.Tag,
		Name:         input.Name,
		Metadata:     input.Metadata,
		CreationDate: timestamppb.New(s.GetTimeService().GetTransactionTime()),
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
	s.GetLogger().Debug().Str("requester", requester).Interface("computePlanUpdate", a).Msg("Updating compute plan")
	err := a.Validate()
	if err != nil {
		return orcerrors.FromValidationError(asset.ComputePlanKind, err)
	}

	plan, err := s.GetComputePlanDBAL().GetComputePlan(a.Key)
	if err != nil {
		return orcerrors.NewNotFound(asset.ComputePlanKind, a.Key)
	}

	if requester != plan.Owner {
		return orcerrors.NewPermissionDenied("requester does not own the compute plan")
	}

	// Update compute plan name
	plan.Name = a.Name

	err = s.GetComputePlanDBAL().SetComputePlanName(plan, plan.Name)
	if err != nil {
		return err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  plan.Key,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		Asset:     &asset.Event_ComputePlan{ComputePlan: plan},
	}
	return s.GetEventService().RegisterEvents(event)
}

func (s *ComputePlanService) failPlan(key string) error {
	plan, err := s.GetPlan(key)
	if err != nil {
		return err
	}

	if plan.IsTerminated() {
		return orcerrors.NewTerminatedComputePlan(plan.Key)
	}

	failureDate := s.GetTimeService().GetTransactionTime()
	return s.GetComputePlanDBAL().FailComputePlan(plan, failureDate)
}

func (s *ComputePlanService) cancelPlan(plan *asset.ComputePlan) error {
	if plan.IsTerminated() {
		return orcerrors.NewTerminatedComputePlan(plan.Key)
	}

	cancelationDate := s.GetTimeService().GetTransactionTime()
	err := s.GetComputePlanDBAL().CancelComputePlan(plan, cancelationDate)
	if err != nil {
		return err
	}

	plan.CancelationDate = timestamppb.New(cancelationDate)

	event := &asset.Event{
		AssetKey:  plan.Key,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		Asset:     &asset.Event_ComputePlan{ComputePlan: plan},
	}

	return s.GetEventService().RegisterEvents(event)
}

func (s *ComputePlanService) computePlanExists(key string) (bool, error) {
	return s.GetComputePlanDBAL().ComputePlanExists(key)
}

// IsPlanRunning indicates whether there are tasks belonging to the compute plan
// being executed or waiting to be executed
func (s *ComputePlanService) IsPlanRunning(key string) (bool, error) {
	plan, err := s.GetPlan(key)
	if plan.IsTerminated() {
		return false, err
	} else {
		return s.GetComputePlanDBAL().IsPlanRunning(key)
	}
}
