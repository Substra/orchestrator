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

	"github.com/go-playground/log/v7"
	"github.com/looplab/fsm"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orchestrationErrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
)

// ComputePlanAPI defines the methods to act on ComputePlans
type ComputePlanAPI interface {
	RegisterPlan(plan *asset.NewComputePlan, owner string) (*asset.ComputePlan, error)
	GetPlan(key string) (*asset.ComputePlan, error)
	GetPlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error)
	ApplyPlanAction(key string, action asset.ComputePlanAction, requester string) error
}

// ComputePlanServiceProvider defines an object able to provide a ComputePlanAPI instance
type ComputePlanServiceProvider interface {
	GetComputePlanService() ComputePlanAPI
}

// ComputePlanDependencyProvider defines what the ComputePlanService needs to perform its duty
type ComputePlanDependencyProvider interface {
	persistence.ComputePlanDBALProvider
	persistence.ComputeTaskDBALProvider
	event.QueueProvider
	ComputeTaskServiceProvider
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
	log.WithField("plan", input).WithField("owner", owner).Debug("Registering new compute plan")
	err := input.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", orchestrationErrors.ErrInvalidAsset, err.Error())
	}

	exist, err := s.GetComputePlanDBAL().ComputePlanExists(input.Key)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, fmt.Errorf("plan %s already exists: %w", input.Key, orchestrationErrors.ErrConflict)
	}

	plan := &asset.ComputePlan{
		Key:      input.Key,
		Owner:    owner,
		Tag:      input.Tag,
		Metadata: input.Metadata,
	}

	err = s.GetComputePlanDBAL().AddComputePlan(plan)
	if err != nil {
		return nil, err
	}

	event := event.Event{
		AssetKind: asset.ComputePlanKind,
		AssetID:   plan.Key,
		EventKind: event.AssetCreated,
		Metadata: map[string]string{
			"creator": plan.Owner,
		},
	}
	err = s.GetEventQueue().Enqueue(&event)
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
		return fmt.Errorf("only plan owner can act on it: %w", orchestrationErrors.ErrPermissionDenied)
	}

	switch action {
	case asset.ComputePlanAction_PLAN_ACTION_CANCELED:
		return s.cancelPlan(plan)
	default:
		return fmt.Errorf("plan action unimplemented: %w", orchestrationErrors.ErrUnimplemented)
	}
}

func (s *ComputePlanService) GetPlan(key string) (*asset.ComputePlan, error) {
	return s.GetComputePlanDBAL().GetComputePlan(key)
}

func (s *ComputePlanService) GetPlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error) {
	return s.GetComputePlanDBAL().QueryComputePlans(p)
}

func (s *ComputePlanService) cancelPlan(plan *asset.ComputePlan) error {
	tasks, err := s.GetComputeTaskDBAL().GetComputePlanTasks(plan.Key)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err := s.GetComputeTaskService().ApplyTaskAction(task.Key, asset.ComputeTaskAction_TASK_ACTION_CANCELED, fmt.Sprintf("compute plan %s is cancelled", plan.Key), plan.Owner)
		if _, isInvalidEvent := err.(fsm.InvalidEventError); isInvalidEvent {
			log.WithError(err).WithField("taskKey", task.Key).WithField("taskStatus", task.Status).Debug("skipping task cancellation: expected error")
		} else {
			return err
		}
	}

	return nil
}