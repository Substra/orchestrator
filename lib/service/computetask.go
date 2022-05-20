package service

import (
	"fmt"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/metrics"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/owkin/orchestrator/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ComputeTaskAPI defines the methods to act on ComputeTasks
type ComputeTaskAPI interface {
	RegisterTasks(tasks []*asset.NewComputeTask, owner string) ([]*asset.ComputeTask, error)
	GetTask(key string) (*asset.ComputeTask, error)
	QueryTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error)
	ApplyTaskAction(key string, action asset.ComputeTaskAction, reason string, requester string) error
	// canDisableModels is internal only (exposed only to other services).
	// it will return true if models produced by the task can be disabled
	canDisableModels(key, requester string) (bool, error)
	// applyTaskAction is internal only, it will trigger a task status update.
	applyTaskAction(task *asset.ComputeTask, action taskTransition, reason string) error
}

// ComputeTaskServiceProvider defines an object able to provide a ComputeTaskAPI instance
type ComputeTaskServiceProvider interface {
	GetComputeTaskService() ComputeTaskAPI
}

// ComputeTaskDependencyProvider defines what the ComputeTaskService needs to perform its duty
type ComputeTaskDependencyProvider interface {
	LoggerProvider
	ChannelProvider
	persistence.ComputeTaskDBALProvider
	EventServiceProvider
	AlgoServiceProvider
	DataManagerServiceProvider
	DataSampleServiceProvider
	PermissionServiceProvider
	NodeServiceProvider
	ComputePlanServiceProvider
	ModelServiceProvider
	TimeServiceProvider
}

// ComputeTaskService is the compute task manipulation entry point
type ComputeTaskService struct {
	ComputeTaskDependencyProvider
	// Keep a local cache of algos, plans and tasks to be used in batch import
	algoStore map[string]*asset.Algo
	taskStore map[string]*asset.ComputeTask
	planStore map[string]*asset.ComputePlan
}

// NewComputeTaskService creates a new service
func NewComputeTaskService(provider ComputeTaskDependencyProvider) *ComputeTaskService {
	return &ComputeTaskService{
		ComputeTaskDependencyProvider: provider,
		algoStore:                     make(map[string]*asset.Algo),
		taskStore:                     make(map[string]*asset.ComputeTask),
		planStore:                     make(map[string]*asset.ComputePlan),
	}
}

// QueryTasks returns tasks matching filter
func (s *ComputeTaskService) QueryTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	s.GetLogger().WithField("pagination", p).WithField("filter", filter).Debug("Querying ComputeTasks")

	return s.GetComputeTaskDBAL().QueryComputeTasks(p, filter)
}

// GetTask return a single task
func (s *ComputeTaskService) GetTask(key string) (*asset.ComputeTask, error) {
	s.GetLogger().WithField("key", key).Debug("Get ComputeTask")

	return s.GetComputeTaskDBAL().GetComputeTask(key)
}

// RegisterTasks creates multiple compute tasks
func (s *ComputeTaskService) RegisterTasks(tasks []*asset.NewComputeTask, owner string) ([]*asset.ComputeTask, error) {
	s.GetLogger().WithField("numTasks", len(tasks)).WithField("owner", owner).Debug("Registering new compute tasks")
	if len(tasks) == 0 {
		return nil, orcerrors.NewBadRequest("no task to register")
	}

	for _, newTask := range tasks {
		err := newTask.Validate()
		if err != nil {
			return nil, orcerrors.FromValidationError(asset.ComputeTaskKind, err)
		}
	}

	existingKeys, err := s.getExistingKeys(tasks)
	if err != nil {
		return nil, err
	}
	sortedTasks, err := s.SortTasks(tasks, existingKeys)
	if err != nil {
		return nil, err
	}

	registeredTasks := []*asset.ComputeTask{}
	events := []*asset.Event{}

	for _, newTask := range sortedTasks {
		task, err := s.createTask(newTask, owner)
		if err != nil {
			return nil, err
		}
		metrics.TaskRegisteredTotal.WithLabelValues(s.GetChannel(), task.Category.String()).Inc()
		registeredTasks = append(registeredTasks, task)

		event := &asset.Event{
			AssetKey:  task.Key,
			EventKind: asset.EventKind_EVENT_ASSET_CREATED,
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			Asset:     &asset.Event_ComputeTask{ComputeTask: task},
			Metadata: map[string]string{
				"status": task.Status.String(),
				"worker": task.Worker,
			},
		}
		events = append(events, event)

	}

	err = s.GetComputeTaskDBAL().AddComputeTasks(registeredTasks...)
	if err != nil {
		return nil, err
	}
	err = s.GetEventService().RegisterEvents(events...)
	if err != nil {
		return nil, err
	}

	metrics.TaskRegistrationBatchSize.WithLabelValues(s.GetChannel()).Observe(float64(len(registeredTasks)))

	return registeredTasks, nil
}

// SortTasks is a function to sort a list of tasks in a valid order for their registration.
// It performs a topological sort of the tasks such that for every dependency from task A to B
// A comes before B in the resulting list of tasks.
// A topological ordering is possible only if the graph is a DAG and has no cycles. This function will
// raise an error if there is a cycle in the list of tasks.
// This sorting function is based on Kahn's algorithm.
func (s *ComputeTaskService) SortTasks(newTasks []*asset.NewComputeTask, existingTasks []string) ([]*asset.NewComputeTask, error) {
	sortedTasks := make([]*asset.NewComputeTask, len(newTasks))
	unsortedTasks := make([]*asset.NewComputeTask, len(newTasks))
	copy(unsortedTasks, newTasks)

	unsortedParentsCount := make(map[string]int, len(unsortedTasks))
	tasksWithoutUnsortedDependency := []*asset.NewComputeTask{}

	for i := 0; i < len(unsortedTasks); i++ {
		unsortedParentsCount[unsortedTasks[i].Key] = 0
		// We count the number of parents that are not already registered in the persistence layer
		for _, parent := range unsortedTasks[i].GetParentTaskKeys() {
			if !utils.StringInSlice(existingTasks, parent) {
				unsortedParentsCount[unsortedTasks[i].Key]++
			}
		}

		if unsortedParentsCount[unsortedTasks[i].Key] == 0 {
			tasksWithoutUnsortedDependency = append(tasksWithoutUnsortedDependency, unsortedTasks[i])
			unsortedTasks = append(unsortedTasks[:i], unsortedTasks[i+1:]...)
			i-- // We go back one index as we removed the element at position i
		}
	}

	sortedTasksCount := 0
	for len(tasksWithoutUnsortedDependency) > 0 {
		currentTask := tasksWithoutUnsortedDependency[0]
		tasksWithoutUnsortedDependency = tasksWithoutUnsortedDependency[1:]

		sortedTasks[sortedTasksCount] = currentTask
		sortedTasksCount++

		for i := 0; i < len(unsortedTasks); i++ {
			for _, key := range unsortedTasks[i].ParentTaskKeys {
				if key == currentTask.Key {
					unsortedParentsCount[unsortedTasks[i].Key]--
					if unsortedParentsCount[unsortedTasks[i].Key] == 0 {
						// As it has no remaining dependency the task is ready to be added to the sorted list.
						// We remove the task from the unsorted list to make our slice length decrease over time
						// and avoid going through all the tasks that are already sorted an have no remaining dependencies.
						tasksWithoutUnsortedDependency = append(tasksWithoutUnsortedDependency, unsortedTasks[i])
						unsortedTasks = append(unsortedTasks[:i], unsortedTasks[i+1:]...)
						i-- // We go back one index as we removed the element at position i
					}
				}
			}
		}
	}

	if len(unsortedTasks) > 0 {
		s.GetLogger().
			WithField("unsortedTasks", len(unsortedTasks)).
			WithField("existingTasks", len(existingTasks)).
			Debug("Failed to sort tasks, cyclic dependency in compute plan graph or unknown parent")
		return nil, orcerrors.NewInvalidAsset(fmt.Sprintf("cyclic dependency in compute plan graph or unknown task parent, unsorted_tasks_count: %d", len(unsortedTasks)))
	}

	return sortedTasks, nil
}

// createTask converts a NewComputeTask into a ComputeTask.
// It does not persist nor dispatch events.
func (s *ComputeTaskService) createTask(input *asset.NewComputeTask, owner string) (*asset.ComputeTask, error) {
	computePlan, err := s.getCachedCP(input.ComputePlanKey)
	if err != nil {
		return nil, err
	}

	if computePlan.Owner != owner {
		return nil, orcerrors.NewPermissionDenied("Cannot register tasks to a compute plan you don't own")
	}

	taskExist, err := s.GetComputeTaskDBAL().ComputeTaskExists(input.Key)
	if err != nil {
		return nil, err
	}
	if taskExist {
		return nil, orcerrors.NewConflict(asset.ComputeTaskKind, input.Key)
	}

	parentTasks, err := s.getRegisteredTasks(input.ParentTaskKeys...)
	if err != nil {
		return nil, err
	}
	if !s.isCompatibleWithParents(input.Category, parentTasks) {
		return nil, orcerrors.NewInvalidAsset("incompatible models from parent tasks")
	}

	status := getInitialStatusFromParents(parentTasks)

	if status == asset.ComputeTaskStatus_STATUS_CANCELED || status == asset.ComputeTaskStatus_STATUS_FAILED {
		return nil, orcerrors.NewError(orcerrors.ErrIncompatibleTaskStatus, fmt.Sprintf("cannot create a task with status %q", status.String()))
	}

	if err := s.allModelsAvailable(parentTasks); err != nil {
		return nil, err
	}

	// TODO: validate inputs

	task := &asset.ComputeTask{
		Key:            input.Key,
		Category:       input.Category,
		Owner:          owner,
		ComputePlanKey: input.ComputePlanKey,
		Metadata:       input.Metadata,
		Status:         status,
		Rank:           getRank(parentTasks),
		ParentTaskKeys: input.ParentTaskKeys,
		CreationDate:   timestamppb.New(s.GetTimeService().GetTransactionTime()),
		Inputs:         input.Inputs,
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
		err = orcerrors.NewInvalidAsset(fmt.Sprintf("unknown task data %T", x))
	}
	if err != nil {
		return nil, err
	}

	err = s.checkCanProcessParents(task.Worker, parentTasks, input.Category)
	if err != nil {
		return nil, err
	}

	s.taskStore[task.Key] = task

	return task, nil
}

// Models produced by a task can only be disabled if all those conditions are met:
// - the compute plan has the DeleteIntermediaryModel set
// - task has train children, ie: not at the tip of the compute plan (test children are ignored)
// - task is in a terminal state (done, failed, canceled)
// - all children are in a terminal state
func (s *ComputeTaskService) canDisableModels(key string, requester string) (bool, error) {
	logger := s.GetLogger().WithField("taskKey", key)
	task, err := s.GetTask(key)
	if err != nil {
		return false, err
	}
	if task.Worker != requester {
		return false, orcerrors.NewPermissionDenied("only the worker can disable a task outputs")
	}

	state := newState(&dumbUpdater, task)
	if len(state.AvailableTransitions()) > 0 {
		logger.WithField("status", state.Current()).Debug("skip model disable: task not in final state")
		return false, nil
	}

	planAllowIntermediary, err := s.GetComputePlanService().canDeleteModels(task.ComputePlanKey)
	if err != nil {
		return false, err
	}
	if !planAllowIntermediary {
		logger.WithField("computePlanKey", task.ComputePlanKey).Debug("skip model disable: DeleteIntermediaryModels is false")
		return false, nil
	}

	children, err := s.GetComputeTaskDBAL().GetComputeTaskChildren(key)
	if err != nil {
		return false, err
	}

	trainChildren := 0

	for _, child := range children {
		if child.Category != asset.ComputeTaskCategory_TASK_TEST {
			trainChildren++
		}
		state := newState(&dumbUpdater, child)
		if len(state.AvailableTransitions()) > 0 {
			logger.WithField("childKey", child.Key).Debug("cannot disable model: task has active children")
			return false, nil
		}
	}

	if trainChildren == 0 {
		logger.Debug("cannot disable model: task has no children")
		return false, nil
	}

	return true, nil
}

// getCheckedAlgo returns the Algo identified by given key,
// it will return an error if the algorithm is not processable by the owner or not compatible with the task.
func (s *ComputeTaskService) getCheckedAlgo(algoKey string, owner string, taskCategory asset.ComputeTaskCategory) (*asset.Algo, error) {
	if _, ok := s.algoStore[algoKey]; !ok {
		algo, err := s.GetAlgoService().GetAlgo(algoKey)
		if err != nil {
			return nil, err
		}
		s.algoStore[algoKey] = algo
	}
	algo := s.algoStore[algoKey]

	canProcess := s.GetPermissionService().CanProcess(algo.Permissions, owner)
	if !canProcess {
		return nil, orcerrors.NewPermissionDenied(fmt.Sprintf("not authorized to process algo %q", algo.Key))
	}

	if !isAlgoCompatible(taskCategory, algo.Category) {
		return nil, orcerrors.NewInvalidAsset("algo category is not compatible with task category")
	}

	return algo, nil
}

// getCheckedDataManager returns the DataManager identified by the given key,
// it will return an error if the DataManager is not processable by owner or DataSamples don't share the common manager.
func (s *ComputeTaskService) getCheckedDataManager(key string, dataSampleKeys []string, owner string) (*asset.DataManager, error) {
	datamanager, err := s.GetDataManagerService().GetDataManager(key)
	if err != nil {
		return nil, err
	}
	canProcess := s.GetPermissionService().CanProcess(datamanager.Permissions, owner)
	if !canProcess {
		return nil, orcerrors.NewPermissionDenied(fmt.Sprintf("not authorized to process datamanager %q", datamanager.Key))
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
		return orcerrors.NewInvalidAsset("cannot create task with test data")
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
	task.LogsPermission = datamanager.LogsPermission

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

	modelPermissions, err := s.GetPermissionService().CreatePermissions(task.Owner, &asset.NewPermissions{Public: false})
	if err != nil {
		return err
	}

	logsPermission, err := s.GetPermissionService().CreatePermission(task.Owner, &asset.NewPermissions{Public: false})
	if err != nil {
		return err
	}

	for _, p := range parentTasks {
		var modelPerms *asset.Permissions
		switch p.Data.(type) {
		case *asset.ComputeTask_Composite:
			modelPerms = p.Data.(*asset.ComputeTask_Composite).Composite.TrunkPermissions
		case *asset.ComputeTask_Aggregate:
			modelPerms = p.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions
		case *asset.ComputeTask_Train:
			modelPerms = p.Data.(*asset.ComputeTask_Train).Train.ModelPermissions
		default:
			return orcerrors.NewPermissionDenied(fmt.Sprintf("cannot process parent task %q", p.Key))
		}
		modelPermissions = s.GetPermissionService().UnionPermissions(modelPerms, modelPermissions)

		logsPermission = s.GetPermissionService().UnionPermission(p.LogsPermission, logsPermission)
	}

	taskData := &asset.AggregateTrainTaskData{
		ModelPermissions: modelPermissions,
	}
	task.Data = &asset.ComputeTask_Aggregate{
		Aggregate: taskData,
	}
	task.Worker = node.Id
	task.Algo = algo
	task.LogsPermission = logsPermission

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
		return orcerrors.NewInvalidAsset("cannot create task with test data")
	}

	algo, err := s.getCheckedAlgo(taskInput.AlgoKey, task.Owner, task.Category)
	if err != nil {
		return err
	}

	permissions := s.GetPermissionService().IntersectPermissions(algo.Permissions, datamanager.Permissions)

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
	task.LogsPermission = datamanager.LogsPermission

	return nil
}

// setTestData hydrates task specific TestTaskData from input
func (s *ComputeTaskService) setTestData(input *asset.NewTestTaskData, task *asset.ComputeTask, parentTasks []*asset.ComputeTask) error {
	datamanager, err := s.getCheckedDataManager(input.DataManagerKey, input.DataSampleKeys, task.Owner)
	if err != nil {
		return err
	}

	for _, metricKey := range input.MetricKeys {
		metricExists, err := s.GetAlgoService().AlgoExists(metricKey)
		if err != nil {
			return err
		}
		if !metricExists {
			return orcerrors.NewNotFound(asset.MetricKind, metricKey)
		}
		// ensure the task will be able to download the metric
		ok, err := s.GetAlgoService().CanDownload(metricKey, datamanager.Owner)
		if err != nil {
			return err
		}
		if !ok {
			return orcerrors.NewPermissionDenied(fmt.Sprintf("datamanager owner cannot download the metric %q", metricKey))
		}
	}

	taskData := &asset.TestTaskData{
		DataManagerKey: input.DataManagerKey,
		DataSampleKeys: input.DataSampleKeys,
		MetricKeys:     input.MetricKeys,
	}

	task.Data = &asset.ComputeTask_Test{
		Test: taskData,
	}
	task.Worker = datamanager.Owner
	task.LogsPermission = datamanager.LogsPermission

	// Should not happen since it is validated by the NewTrain
	if len(parentTasks) != 1 {
		return orcerrors.NewInvalidAsset("invalid number of parents")
	}
	task.Algo = parentTasks[0].Algo
	task.ComputePlanKey = parentTasks[0].ComputePlanKey
	// In case of test tasks there is only one parent (see isCompatibleWithParents)
	// and the test task should have the same rank
	task.Rank = parentTasks[0].Rank

	return nil
}

// checkCanProcessParents raises an error if one of the parent is not processable
func (s *ComputeTaskService) checkCanProcessParents(worker string, parentTasks []*asset.ComputeTask, category asset.ComputeTaskCategory) error {
	switch category {
	case asset.ComputeTaskCategory_TASK_AGGREGATE, asset.ComputeTaskCategory_TASK_TEST, asset.ComputeTaskCategory_TASK_TRAIN:
		return s.checkGenericCanProcessParents(worker, parentTasks, category)
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		return s.checkCompositeCanProcessParents(worker, parentTasks)
	default:
		return orcerrors.NewUnimplemented("invalid task category")
	}
}

// checkGenericCanProcessParents will loop over parents, regardless of task category and return an error if there are insufficient permissions.
func (s *ComputeTaskService) checkGenericCanProcessParents(worker string, parentTasks []*asset.ComputeTask, category asset.ComputeTaskCategory) error {
	for _, p := range parentTasks {
		switch p.Data.(type) {
		case *asset.ComputeTask_Composite:
			trunkPerms := p.Data.(*asset.ComputeTask_Composite).Composite.TrunkPermissions
			if !s.GetPermissionService().CanProcess(trunkPerms, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process composite parent task %q, worker %q is not allowed to process trunk model by permissions %v", p.Key, worker, trunkPerms,
				))
			}
			headPerms := p.Data.(*asset.ComputeTask_Composite).Composite.HeadPermissions
			if category == asset.ComputeTaskCategory_TASK_TEST && !s.GetPermissionService().CanProcess(headPerms, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process composite parent task %q, worker %q is not allowed to process head model by permissions %v", p.Key, worker, headPerms,
				))
			}
		case *asset.ComputeTask_Aggregate:
			permissions := p.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions
			if !s.GetPermissionService().CanProcess(permissions, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process aggregate parent task %q, worker %q is not allowed to process model by permissions %v", p.Key, worker, permissions,
				))
			}
		case *asset.ComputeTask_Train:
			permissions := p.Data.(*asset.ComputeTask_Train).Train.ModelPermissions
			if !s.GetPermissionService().CanProcess(permissions, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process train parent task %q, worker %q is not allowed to process model by permissions %v", p.Key, worker, permissions,
				))
			}
		default:
			return orcerrors.NewUnimplemented("invalid parent category")
		}
	}

	return nil
}

// checkCompositeCanProcessParents returns an error if a composite task with given parents has insufficient permissions
// It depends on both parent category and order if there are two composite parents.
// The first composite parent will provide the HEAD model, while the second will provide the TRUNK
func (s *ComputeTaskService) checkCompositeCanProcessParents(worker string, parentTasks []*asset.ComputeTask) error {
	// compositeInputs contains a couple of tasks: first one will be checked for HEAD perm, second one for TRUNK perm
	compositeInputs := make([]*asset.ComputeTask, 0, 2)
	compositeInputs = append(compositeInputs, parentTasks...)

	if len(parentTasks) == 1 {
		// Single composite parent: it should be checked for both head and trunk permissions
		compositeInputs = append(compositeInputs, parentTasks[0])
	}

	hasAggregateParent := false
	for _, p := range parentTasks {
		if _, ok := p.Data.(*asset.ComputeTask_Aggregate); ok {
			hasAggregateParent = true
		}
	}

	for i, p := range compositeInputs {
		switch p.Data.(type) {
		case *asset.ComputeTask_Composite:
			switch {
			case hasAggregateParent || i == 0:
				// If there is an aggregate parent, the HEAD come from the composite parent, regardless of parent ordering.
				// If there is no aggregate parent, the first composite parent should contribute the HEAD model.
				headPerms := p.Data.(*asset.ComputeTask_Composite).Composite.HeadPermissions
				if !s.GetPermissionService().CanProcess(headPerms, worker) {
					return orcerrors.NewPermissionDenied(fmt.Sprintf(
						"cannot process composite parent task %q, worker %q is not allowed to process head model by permissions %v", p.Key, worker, headPerms,
					))
				}
			case !hasAggregateParent || i == 1:
				// Without aggregate parent, the second composite parent should contribute the trunk model.
				trunkPerms := p.Data.(*asset.ComputeTask_Composite).Composite.TrunkPermissions
				if !s.GetPermissionService().CanProcess(trunkPerms, worker) {
					return orcerrors.NewPermissionDenied(fmt.Sprintf(
						"cannot process composite parent task %q, worker %q is not allowed to process trunk model by permissions %v", p.Key, worker, trunkPerms,
					))
				}
			}
		case *asset.ComputeTask_Aggregate:
			permissions := p.Data.(*asset.ComputeTask_Aggregate).Aggregate.ModelPermissions
			if !s.GetPermissionService().CanProcess(permissions, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process aggregate parent task %q, worker %q is not allowed to process model by permissions %v", p.Key, worker, permissions,
				))
			}
		default:
			return orcerrors.NewUnimplemented("invalid parent category")
		}
	}

	return nil
}

// getRegisteredTask will return the tasks from the current batch or the database if not found.
// The tasks are returned in the same order as the keys.
func (s *ComputeTaskService) getRegisteredTasks(keys ...string) ([]*asset.ComputeTask, error) {
	bag := make(map[string]*asset.ComputeTask)
	notInStore := []string{}

	for _, key := range keys {
		if task, ok := s.taskStore[key]; ok {
			bag[key] = task
		} else {
			notInStore = append(notInStore, key)
		}
	}

	if len(notInStore) > 0 {
		tasks, err := s.GetComputeTaskDBAL().GetComputeTasks(notInStore)

		if err != nil {
			return nil, err
		}

		for _, task := range tasks {
			bag[task.Key] = task
		}
	}

	result := make([]*asset.ComputeTask, len(keys))

	// Add the tasks in order
	for i, k := range keys {
		result[i] = bag[k]
	}

	return result, nil
}

// getExistingKeys returns the list of tasks already persisted.
func (s *ComputeTaskService) getExistingKeys(tasks []*asset.NewComputeTask) ([]string, error) {
	parents := []string{}

	for _, task := range tasks {
		parents = append(parents, task.ParentTaskKeys...)
	}

	existingKeys, err := s.GetComputeTaskDBAL().GetExistingComputeTaskKeys(parents)
	if err != nil {
		return nil, err
	}

	return existingKeys, nil
}

// isCompatibleWithParents checks task compatibility with parents tasks
func (s *ComputeTaskService) isCompatibleWithParents(category asset.ComputeTaskCategory, parents []*asset.ComputeTask) bool {
	inputs := map[asset.ComputeTaskCategory]uint32{}

	for _, p := range parents {
		inputs[p.Category]++
	}

	s.GetLogger().WithField("category", category).WithField("parents", inputs).Debug("checking parent compatibility")

	noTest := inputs[asset.ComputeTaskCategory_TASK_TEST] == 0
	noTrain := inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 0
	noComposite := inputs[asset.ComputeTaskCategory_TASK_COMPOSITE] == 0
	noParent := inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 0
	compositeOnly := inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 0 && inputs[asset.ComputeTaskCategory_TASK_COMPOSITE] == 1
	compositeAndAggregate := inputs[asset.ComputeTaskCategory_TASK_AGGREGATE] == 1 && inputs[asset.ComputeTaskCategory_TASK_COMPOSITE] == 1
	twoComposites := inputs[asset.ComputeTaskCategory_TASK_COMPOSITE] == 2 && inputs[asset.ComputeTaskCategory_TASK_AGGREGATE] == 0

	switch category {
	case asset.ComputeTaskCategory_TASK_TRAIN:
		return noTest && noComposite
	case asset.ComputeTaskCategory_TASK_TEST:
		return noTest && inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 1
	case asset.ComputeTaskCategory_TASK_AGGREGATE:
		return noTest && inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] >= 1
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		return noTest && noTrain && (noParent || compositeOnly || compositeAndAggregate || twoComposites)
	default:
		return false
	}
}

// allModelsAvailable checks that all parent models are available to the task
func (s *ComputeTaskService) allModelsAvailable(parents []*asset.ComputeTask) error {
	for _, p := range parents {
		if p.Status == asset.ComputeTaskStatus_STATUS_DONE {
			models, err := s.GetModelService().GetComputeTaskOutputModels(p.Key)
			if err != nil {
				return err
			}

			for _, m := range models {
				if m.Address == nil {
					return orcerrors.NewInvalidAsset(fmt.Sprintf("Model %q has been disabled", m.Key))
				}
			}
		}
	}

	return nil
}

// getCachedCP returns a cached version of a compute plan
// we cache the result to avoid multiple dbal queries on batch registration
func (s *ComputeTaskService) getCachedCP(key string) (*asset.ComputePlan, error) {
	if _, ok := s.planStore[key]; !ok {
		plan, err := s.GetComputePlanService().GetPlan(key)
		if err != nil {
			return nil, err
		}
		s.planStore[key] = plan
	}
	return s.planStore[key], nil
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

// getRank determines the rank of a task from its parents.
// A task with no parents has a rank of 0.
// Otherwise its rank is set to max(parentRanks) + 1.
func getRank(parentTasks []*asset.ComputeTask) int32 {
	if len(parentTasks) == 0 {
		return 0
	}

	maxParentRank := int32(0)
	for _, p := range parentTasks {
		if p.Rank > maxParentRank {
			maxParentRank = p.Rank
		}
	}

	return maxParentRank + 1
}
