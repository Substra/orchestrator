// Copyright 2020 Owkin Inc.
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
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	orchestrationErrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/owkin/orchestrator/utils"
)

// ComputeTaskAPI defines the methods to act on ComputeTasks
type ComputeTaskAPI interface {
	// RegisterTask creates a new ComputeTask
	RegisterTask(task *asset.NewComputeTask, owner string) (*asset.ComputeTask, error)
	GetTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error)
	ApplyTaskAction(key string, action asset.ComputeTaskAction, reason string) error
}

// ComputeTaskServiceProvider defines an object able to provide a ComputeTaskAPI instance
type ComputeTaskServiceProvider interface {
	GetComputeTaskService() ComputeTaskAPI
}

// ComputeTaskDependencyProvider defines what the ComputeTaskService needs to perform its duty
type ComputeTaskDependencyProvider interface {
	persistence.ComputeTaskDBALProvider
	event.QueueProvider
	AlgoServiceProvider
	DataManagerServiceProvider
	DataSampleServiceProvider
	PermissionServiceProvider
	ObjectiveServiceProvider
	NodeServiceProvider
}

// ComputeTaskService is the compute task manipulation entry point
type ComputeTaskService struct {
	ComputeTaskDependencyProvider
}

// NewComputeTaskService creates a new service
func NewComputeTaskService(provider ComputeTaskDependencyProvider) *ComputeTaskService {
	return &ComputeTaskService{provider}
}

// GetTasks returns tasks matching filter
func (s *ComputeTaskService) GetTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	log.WithField("pagination", p).WithField("filter", filter).Debug("Querying ComputeTasks")

	return s.GetComputeTaskDBAL().QueryComputeTasks(p, filter)
}

// RegisterTask creates a new ComputeTask
func (s *ComputeTaskService) RegisterTask(input *asset.NewComputeTask, owner string) (*asset.ComputeTask, error) {
	log.WithField("task", input).WithField("owner", owner).Debug("Registering new compute task")
	err := input.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", orchestrationErrors.ErrInvalidAsset, err.Error())
	}

	// TODO: compute plan should exist

	taskExist, err := s.GetComputeTaskDBAL().ComputeTaskExists(input.Key)
	if err != nil {
		return nil, err
	}
	if taskExist {
		return nil, fmt.Errorf("task %s already exists: %w", input.Key, orchestrationErrors.ErrConflict)
	}

	parentTasks, err := s.GetComputeTaskDBAL().GetComputeTasks(input.ParentTaskKeys)
	if err != nil {
		return nil, err
	}
	if !isParentCompatible(input.Category, parentTasks) {
		return nil, fmt.Errorf("incompatible models from parent tasks: %w", orchestrationErrors.ErrInvalidAsset)
	}

	err = s.checkCanProcessParents(owner, parentTasks, input.Category)
	if err != nil {
		return nil, err
	}

	status := getInitialStatusFromParents(parentTasks)

	if status == asset.ComputeTaskStatus_STATUS_CANCELED || status == asset.ComputeTaskStatus_STATUS_FAILED {
		return nil, fmt.Errorf("cannot create a task with status %s: %w", status, orchestrationErrors.ErrIncompatibleTaskStatus)
	}

	task := &asset.ComputeTask{
		Key:            input.Key,
		Category:       input.Category,
		Owner:          owner,
		ComputePlanKey: input.ComputePlanKey,
		Metadata:       input.Metadata,
		Status:         status,
		Rank:           input.Rank,
		ParentTaskKeys: input.ParentTaskKeys,
	}

	switch x := input.Data.(type) {
	case *asset.NewComputeTask_Composite:
		err = s.setCompositeData(input, input.Data.(*asset.NewComputeTask_Composite).Composite, task)
	case *asset.NewComputeTask_Aggregate:
		err = s.setAggregateData(input, input.Data.(*asset.NewComputeTask_Aggregate).Aggregate, task, parentTasks)
	case *asset.NewComputeTask_Test:
		err = s.setTestData(input.Data.(*asset.NewComputeTask_Test).Test, task, parentTasks)
	case *asset.NewComputeTask_Train:
		err = s.setTrainData(input, input.Data.(*asset.NewComputeTask_Train).Train, task)
	default:
		// Should never happen, validated above
		err = fmt.Errorf("unkwown task data: %T, %w", x, errors.ErrInvalidAsset)
	}
	if err != nil {
		return nil, err
	}

	err = s.GetComputeTaskDBAL().AddComputeTask(task)
	if err != nil {
		return nil, err
	}

	event := event.Event{
		AssetKind: asset.ComputeTaskKind,
		AssetID:   task.Key,
		EventKind: event.AssetCreated,
		Metadata: map[string]string{
			"status": task.Status.String(),
		},
	}

	err = s.GetEventQueue().Enqueue(&event)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// getCheckedAlgo returns the Algo identified by given key,
// it will return an error if the algorithm is not processable by the owner or not compatible with the task.
func (s *ComputeTaskService) getCheckedAlgo(algoKey string, owner string, taskCategory asset.ComputeTaskCategory) (*asset.Algo, error) {
	algo, err := s.GetAlgoService().GetAlgo(algoKey)
	if err != nil {
		return nil, err
	}
	canProcess := s.GetPermissionService().CanProcess(algo.Permissions, owner)
	if !canProcess {
		return nil, fmt.Errorf("not authorized to process algo %s: %w", algo.Key, orchestrationErrors.ErrPermissionDenied)
	}

	if !isAlgoCompatible(taskCategory, algo.Category) {
		return nil, fmt.Errorf("algo category is not compatible with task category: %w", orchestrationErrors.ErrInvalidAsset)
	}

	return algo, nil
}

// getCheckedDataManager returns the DataManager identified by the given key,
// it will return an error if the DataManager is not processable by owner or DataSamples don't share the common manager.
func (s *ComputeTaskService) getCheckedDataManager(key string, dataSampleKeys []string, owner string) (*asset.DataManager, error) {
	datamanager, err := s.GetDataManagerService().GetDataManager(key)
	if err != nil {
		return nil, fmt.Errorf("datamanager not found: %w", orchestrationErrors.ErrReferenceNotFound)
	}
	canProcess := s.GetPermissionService().CanProcess(datamanager.Permissions, owner)
	if !canProcess {
		return nil, fmt.Errorf("not authorized to process datamanager %s: %w", datamanager.Key, orchestrationErrors.ErrPermissionDenied)
	}
	err = s.GetDataSampleService().CheckSameManager(key, dataSampleKeys)
	if err != nil {
		return nil, err
	}

	return datamanager, err
}

// setCompositeData hydrates task specific CompositeTrainTaskData from input
func (s *ComputeTaskService) setCompositeData(taskInput *asset.NewComputeTask, specificInput *asset.NewCompositeTrainTaskData, task *asset.ComputeTask) error {
	datamanager, err := s.getCheckedDataManager(specificInput.DataManagerKey, specificInput.DataSampleKeys, task.Owner)
	if err != nil {
		return err
	}

	hasTest, err := s.GetDataSampleService().ContainsTestSample(specificInput.DataSampleKeys)
	if err != nil {
		return err
	}
	if hasTest {
		return fmt.Errorf("cannot create task with test data: %w", orchestrationErrors.ErrInvalidAsset)
	}

	algo, err := s.getCheckedAlgo(taskInput.AlgoKey, task.Owner, task.Category)
	if err != nil {
		return err
	}

	headPermissions, err := s.GetPermissionService().CreatePermissions(datamanager.Owner, nil)
	if err != nil {
		return err
	}
	trunkPermissions, err := s.GetPermissionService().CreatePermissions(datamanager.Owner, specificInput.TrunkPermissions)
	if err != nil {
		return err
	}

	taskData := &asset.CompositeTrainTaskData{
		DataManagerKey:   datamanager.Key,
		DataSampleKeys:   specificInput.DataSampleKeys,
		HeadPermissions:  headPermissions,
		TrunkPermissions: trunkPermissions,
	}

	task.Data = &asset.ComputeTask_Composite{
		Composite: taskData,
	}
	task.Worker = datamanager.Owner
	task.Algo = algo

	return nil
}

// setAggregateData hydrates task specific AggregateTrainTaskData from input
func (s *ComputeTaskService) setAggregateData(taskInput *asset.NewComputeTask, input *asset.NewAggregateTrainTaskData, task *asset.ComputeTask, parentTasks []*asset.ComputeTask) error {
	node, err := s.GetNodeService().GetNode(input.Worker)
	if err != nil {
		return err
	}
	algo, err := s.getCheckedAlgo(taskInput.AlgoKey, task.Owner, task.Category)
	if err != nil {
		return err
	}

	perms, err := s.GetPermissionService().CreatePermissions(task.Owner, &asset.NewPermissions{Public: true})
	if err != nil {
		return err
	}

	for _, p := range parentTasks {
		var permissions *asset.Permissions
		switch p.Data.(type) {
		case *asset.ComputeTask_Composite:
			permissions = p.Data.(*asset.ComputeTask_Composite).Composite.TrunkPermissions
		case *asset.ComputeTask_Aggregate:
			permissions = p.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions
		case *asset.ComputeTask_Train:
			permissions = p.Data.(*asset.ComputeTask_Train).Train.ModelPermissions
		default:
			return fmt.Errorf("cannot process parent task %s: %w", p.Key, errors.ErrPermissionDenied)
		}
		perms = s.GetPermissionService().MergePermissions(permissions, perms)
	}

	taskData := &asset.AggregateTrainTaskData{
		ModelPermissions: perms,
	}
	task.Data = &asset.ComputeTask_Aggregate{
		Aggregate: taskData,
	}
	task.Worker = node.Id
	task.Algo = algo

	return nil
}

// setTrainData hydrates task specific TrainTaskData from input
func (s *ComputeTaskService) setTrainData(taskInput *asset.NewComputeTask, specificInput *asset.NewTrainTaskData, task *asset.ComputeTask) error {
	datamanager, err := s.getCheckedDataManager(specificInput.DataManagerKey, specificInput.DataSampleKeys, task.Owner)
	if err != nil {
		return err
	}

	hasTest, err := s.GetDataSampleService().ContainsTestSample(specificInput.DataSampleKeys)
	if err != nil {
		return err
	}
	if hasTest {
		return fmt.Errorf("cannot create task with test data: %w", orchestrationErrors.ErrInvalidAsset)
	}

	algo, err := s.getCheckedAlgo(taskInput.AlgoKey, task.Owner, task.Category)
	if err != nil {
		return err
	}

	permissions := s.GetPermissionService().MergePermissions(algo.Permissions, datamanager.Permissions)

	taskData := &asset.TrainTaskData{
		DataManagerKey:   datamanager.Key,
		DataSampleKeys:   specificInput.DataSampleKeys,
		ModelPermissions: permissions,
	}

	task.Data = &asset.ComputeTask_Train{
		Train: taskData,
	}
	task.Worker = datamanager.Owner

	task.Algo = algo

	return nil
}

// setTestData hydrates task specific TestTaskData from input
func (s *ComputeTaskService) setTestData(input *asset.NewTestTaskData, task *asset.ComputeTask, parentTasks []*asset.ComputeTask) error {
	objective, err := s.GetObjectiveService().GetObjective(input.ObjectiveKey)
	if err != nil {
		return err
	}

	// Test is certified when using objective test data
	certified := true
	dataManagerKey := objective.DataManagerKey
	dataSampleKeys := objective.DataSampleKeys
	datamanager, err := s.GetDataManagerService().GetDataManager(objective.DataManagerKey)
	if err != nil {
		return fmt.Errorf("datamanager not found: %w", orchestrationErrors.ErrReferenceNotFound)
	}

	if input.DataManagerKey != "" {
		datamanager, err = s.getCheckedDataManager(input.DataManagerKey, input.DataSampleKeys, task.Owner)
		if err != nil {
			return err
		}

		dataManagerKey = input.DataManagerKey
		dataSampleKeys = input.DataSampleKeys

		certified = input.DataManagerKey == objective.DataManagerKey && utils.IsEqual(input.DataSampleKeys, objective.DataSampleKeys)
	}

	taskData := &asset.TestTaskData{
		DataManagerKey: dataManagerKey,
		DataSampleKeys: dataSampleKeys,
		Certified:      certified,
		ObjectiveKey:   objective.Key,
	}

	task.Data = &asset.ComputeTask_Test{
		Test: taskData,
	}
	task.Worker = datamanager.Owner

	// Should not happen since it is validated by the NewTrain
	if len(parentTasks) != 1 {
		return fmt.Errorf("Invalid number of parents: %w", orchestrationErrors.ErrInvalidAsset)
	}
	task.Algo = parentTasks[0].Algo
	task.ComputePlanKey = parentTasks[0].ComputePlanKey
	task.Rank = parentTasks[0].Rank

	return nil
}

// checkCanProcessParents raises an error if one of the parent is not processable
func (s *ComputeTaskService) checkCanProcessParents(requester string, parentTasks []*asset.ComputeTask, category asset.ComputeTaskCategory) error {
	for _, p := range parentTasks {
		switch p.Data.(type) {
		case *asset.ComputeTask_Composite:
			trunkPerms := p.Data.(*asset.ComputeTask_Composite).Composite.TrunkPermissions
			if !s.GetPermissionService().CanProcess(trunkPerms, requester) {
				return fmt.Errorf("cannot process parent task %s: %w", p.Key, errors.ErrPermissionDenied)
			}
			headPerms := p.Data.(*asset.ComputeTask_Composite).Composite.HeadPermissions
			if (category == asset.ComputeTaskCategory_TASK_COMPOSITE || category == asset.ComputeTaskCategory_TASK_TEST) && !s.GetPermissionService().CanProcess(headPerms, requester) {
				return fmt.Errorf("cannot process parent task %s: %w", p.Key, errors.ErrPermissionDenied)
			}
		case *asset.ComputeTask_Aggregate:
			permissions := p.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions
			if !s.GetPermissionService().CanProcess(permissions, requester) {
				return fmt.Errorf("cannot process parent task %s: %w", p.Key, errors.ErrPermissionDenied)
			}
		case *asset.ComputeTask_Train:
			permissions := p.Data.(*asset.ComputeTask_Train).Train.ModelPermissions
			if !s.GetPermissionService().CanProcess(permissions, requester) {
				return fmt.Errorf("cannot process parent task %s: %w", p.Key, errors.ErrPermissionDenied)
			}
		default:
			return fmt.Errorf("cannot process parent task %s: %w", p.Key, errors.ErrPermissionDenied)
		}
	}

	return nil
}

// Check task compatibility with parents tasks
func isParentCompatible(category asset.ComputeTaskCategory, parents []*asset.ComputeTask) bool {
	inputs := map[asset.ComputeTaskCategory]uint32{}

	for _, p := range parents {
		inputs[p.Category]++
	}

	noTest := inputs[asset.ComputeTaskCategory_TASK_TEST] == 0
	noTrain := inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 0
	noAggregate := inputs[asset.ComputeTaskCategory_TASK_AGGREGATE] == 0
	noComposite := inputs[asset.ComputeTaskCategory_TASK_COMPOSITE] == 0
	noParent := inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 0
	compositeOnly := inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 0 && inputs[asset.ComputeTaskCategory_TASK_COMPOSITE] == 1
	compositeAndAggregate := inputs[asset.ComputeTaskCategory_TASK_AGGREGATE] == 1 && inputs[asset.ComputeTaskCategory_TASK_COMPOSITE] == 1

	switch category {
	case asset.ComputeTaskCategory_TASK_TRAIN:
		return noTest && noComposite
	case asset.ComputeTaskCategory_TASK_TEST:
		return noTest && noAggregate && inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 1
	case asset.ComputeTaskCategory_TASK_AGGREGATE:
		return noTest && inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] >= 1
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		return noTest && noTrain && (noParent || compositeOnly || compositeAndAggregate)
	default:
		return false
	}
}

// isAlgoCompatible checks if the given algorithm has an appropriate category wrt taskCategory
func isAlgoCompatible(taskCategory asset.ComputeTaskCategory, algoCategory asset.AlgoCategory) bool {
	switch taskCategory {
	case asset.ComputeTaskCategory_TASK_AGGREGATE:
		return algoCategory == asset.AlgoCategory_ALGO_AGGREGATE
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		return algoCategory == asset.AlgoCategory_ALGO_COMPOSITE
	case asset.ComputeTaskCategory_TASK_TEST:
		return true
	case asset.ComputeTaskCategory_TASK_TRAIN:
		return algoCategory == asset.AlgoCategory_ALGO_SIMPLE
	default:
		// should not happen, that means we probably don't known how to deal with task/algo couple
		return false
	}
}
