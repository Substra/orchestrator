package service

import (
	"fmt"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/metrics"
	"github.com/substra/orchestrator/lib/persistence"
	"github.com/substra/orchestrator/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Compatibility constants to support legacy structures (task.data.aggregate, task.data.composite, etc)
var (
	taskIOModel       = "model"
	taskIOShared      = "shared"
	taskIOLocal       = "local"
	taskIOPredictions = "predictions"
)

// Task statuses in which the inputs are defined
var inputDefinedStatus = []asset.ComputeTaskStatus{
	asset.ComputeTaskStatus_STATUS_DOING,
	asset.ComputeTaskStatus_STATUS_TODO,
}

type namedAlgoOutputs = map[string]*asset.AlgoOutput

// ComputeTaskAPI defines the methods to act on ComputeTasks
type ComputeTaskAPI interface {
	RegisterTasks(tasks []*asset.NewComputeTask, owner string) ([]*asset.ComputeTask, error)
	GetTask(key string) (*asset.ComputeTask, error)
	QueryTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error)
	ApplyTaskAction(key string, action asset.ComputeTaskAction, reason string, requester string) error
	GetInputAssets(key string) ([]*asset.ComputeTaskInputAsset, error)
	// canDisableModels is internal only (exposed only to other services).
	// it will return true if models produced by the task can be disabled
	canDisableModels(key, requester string) (bool, error)
	// applyTaskAction is internal only, it will trigger a task status update.
	applyTaskAction(task *asset.ComputeTask, action taskTransition, reason string) error
	addComputeTaskOutputAsset(output *asset.ComputeTaskOutputAsset) error
	getTaskOutputCounter(taskKey string) (persistence.ComputeTaskOutputCounter, error)
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
	OrganizationServiceProvider
	ComputePlanServiceProvider
	ModelServiceProvider
	TimeServiceProvider
}

// ComputeTaskService is the compute task manipulation entry point
type ComputeTaskService struct {
	ComputeTaskDependencyProvider
	// Keep a local cache of algos, plans and tasks to be used in batch import
	algoStore        map[string]*asset.Algo
	taskStore        map[string]*asset.ComputeTask
	planStore        map[string]*asset.ComputePlan
	dataManagerStore map[string]*asset.DataManager
}

// NewComputeTaskService creates a new service
func NewComputeTaskService(provider ComputeTaskDependencyProvider) *ComputeTaskService {
	return &ComputeTaskService{
		ComputeTaskDependencyProvider: provider,
		algoStore:                     make(map[string]*asset.Algo),
		taskStore:                     make(map[string]*asset.ComputeTask),
		planStore:                     make(map[string]*asset.ComputePlan),
		dataManagerStore:              make(map[string]*asset.DataManager),
	}
}

// QueryTasks returns tasks matching filter
func (s *ComputeTaskService) QueryTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	s.GetLogger().Debug().Interface("pagination", p).Interface("filter", filter).Msg("Querying ComputeTasks")

	return s.GetComputeTaskDBAL().QueryComputeTasks(p, filter)
}

// GetTask return a single task
func (s *ComputeTaskService) GetTask(key string) (*asset.ComputeTask, error) {
	s.GetLogger().Debug().Str("key", key).Msg("Get ComputeTask")

	return s.GetComputeTaskDBAL().GetComputeTask(key)
}

// RegisterTasks creates multiple compute tasks
func (s *ComputeTaskService) RegisterTasks(tasks []*asset.NewComputeTask, owner string) ([]*asset.ComputeTask, error) {
	s.GetLogger().Debug().Int("numTasks", len(tasks)).Str("owner", owner).Msg("Registering new compute tasks")
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
	if len(existingKeys) > 0 {
		return nil, orcerrors.NewConflict(asset.ComputeTaskKind, existingKeys[0])
	}

	existingParentKeys, err := s.getExistingParentKeys(tasks)
	if err != nil {
		return nil, err
	}

	sortedTasks, err := s.sortTasks(tasks, existingParentKeys)
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

// GetInputAssets returns the assets necessary to process the task.
func (s *ComputeTaskService) GetInputAssets(key string) ([]*asset.ComputeTaskInputAsset, error) {
	task, err := s.GetTask(key)
	if err != nil {
		return nil, err
	}

	if !utils.SliceContains(inputDefinedStatus, task.Status) {
		return nil, orcerrors.NewBadRequest(
			fmt.Sprintf("Task inputs may not be defined in current task status (%q)", task.Status.String()),
		)
	}

	inputAssets := make([]*asset.ComputeTaskInputAsset, 0, len(task.Inputs))

	for _, input := range task.Inputs {
		algoInput, ok := task.Algo.Inputs[input.Identifier]
		if !ok {
			// This should not happen since this is checked on registration
			return nil, orcerrors.NewError(orcerrors.ErrInternal, "missing algo input")
		}

		switch inputRef := input.Ref.(type) {
		case *asset.ComputeTaskInput_AssetKey:
			inputAsset, err := s.getInputAsset(algoInput.Kind, inputRef.AssetKey, input.Identifier)
			if err != nil {
				return nil, err
			}
			inputAssets = append(inputAssets, inputAsset)
		case *asset.ComputeTaskInput_ParentTaskOutput:
			outputs, err := s.GetComputeTaskDBAL().
				GetComputeTaskOutputAssets(
					inputRef.ParentTaskOutput.ParentTaskKey,
					inputRef.ParentTaskOutput.OutputIdentifier,
				)
			if err != nil {
				return nil, err
			}

			for _, output := range outputs {
				inputAsset, err := s.getInputAsset(output.AssetKind, output.AssetKey, input.Identifier)
				if err != nil {
					return nil, err
				}
				inputAssets = append(inputAssets, inputAsset)
			}
		default:
			return nil, orcerrors.NewUnimplemented(fmt.Sprintf("unsupported input type: %T", inputRef))
		}
	}

	return inputAssets, nil
}

// sortTasks is a function to sort a list of tasks in a valid order for their registration.
// It performs a topological sort of the tasks such that for every dependency from task A to B
// A comes before B in the resulting list of tasks.
// A topological ordering is possible only if the graph is a DAG and has no cycles. This function will
// raise an error if there is a cycle in the list of tasks.
// This sorting function is based on Kahn's algorithm.
func (s *ComputeTaskService) sortTasks(newTasks []*asset.NewComputeTask, existingTasks []string) ([]*asset.NewComputeTask, error) {
	sortedTasks := make([]*asset.NewComputeTask, len(newTasks))
	unsortedTasks := make([]*asset.NewComputeTask, len(newTasks))
	copy(unsortedTasks, newTasks)

	unsortedParentsCount := make(map[string]int, len(unsortedTasks))
	tasksWithoutUnsortedDependency := []*asset.NewComputeTask{}

	for i := 0; i < len(unsortedTasks); i++ {
		unsortedParentsCount[unsortedTasks[i].Key] = 0
		// We count the number of parents that are not already registered in the persistence layer
		for _, parent := range unsortedTasks[i].GetParentTaskKeys() {
			if !utils.SliceContains(existingTasks, parent) {
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
		s.GetLogger().Debug().
			Int("unsortedTasks", len(unsortedTasks)).
			Int("existingTasks", len(existingTasks)).
			Msg("Failed to sort tasks, cyclic dependency in compute plan graph or unknown parent")
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

	algo, err := s.getCheckedAlgo(input.AlgoKey, owner, input.Category)
	if err != nil {
		return nil, err
	}

	// TODO: validate inputs

	outputs := make(map[string]*asset.ComputeTaskOutput, len(input.Outputs))
	for identifier, output := range input.Outputs {
		perm, err := s.GetPermissionService().CreatePermissions(owner, output.Permissions)
		if err != nil {
			return nil, err
		}
		outputs[identifier] = &asset.ComputeTaskOutput{
			Permissions: perm,
			Transient:   output.Transient,
		}
	}

	task := &asset.ComputeTask{
		Key:            input.Key,
		Algo:           algo,
		Category:       input.Category,
		Owner:          owner,
		ComputePlanKey: input.ComputePlanKey,
		Metadata:       input.Metadata,
		Status:         status,
		Rank:           getRank(parentTasks),
		ParentTaskKeys: input.ParentTaskKeys,
		CreationDate:   timestamppb.New(s.GetTimeService().GetTransactionTime()),
		Inputs:         input.Inputs,
		Outputs:        outputs,
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
	case *asset.NewComputeTask_Predict:
		err = s.setPredictData(input, input.Data.(*asset.NewComputeTask_Predict).Predict, task)
	default:
		// Should never happen, validated above
		err = orcerrors.NewInvalidAsset(fmt.Sprintf("unknown task data %T", x))
	}
	if err != nil {
		return nil, err
	}

	if err := s.validateInputs(task.Inputs, task.Algo.Inputs, task.Owner, task.Worker); err != nil {
		return nil, err
	}

	if err := s.validateOutputs(task.Key, task.Outputs, task.Algo.Outputs); err != nil {
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
// - task has train children, ie: not at the tip of the compute plan (test/predict children are ignored)
// - task is in a terminal state (done, failed, canceled)
// - all children are in a terminal state
func (s *ComputeTaskService) canDisableModels(key string, requester string) (bool, error) {
	logger := s.GetLogger().With().Str("taskKey", key).Logger()
	task, err := s.GetTask(key)
	if err != nil {
		return false, err
	}
	if task.Worker != requester {
		return false, orcerrors.NewPermissionDenied("only the worker can disable a task outputs")
	}

	state := newState(&dumbUpdater, task)
	if len(state.AvailableTransitions()) > 0 {
		logger.Debug().Str("status", state.Current()).Msg("skip model disable: task not in final state")
		return false, nil
	}

	planAllowIntermediary, err := s.GetComputePlanService().canDeleteModels(task.ComputePlanKey)
	if err != nil {
		return false, err
	}
	if !planAllowIntermediary {
		logger.Debug().Str("computePlanKey", task.ComputePlanKey).Msg("skip model disable: DeleteIntermediaryModels is false")
		return false, nil
	}

	children, err := s.GetComputeTaskDBAL().GetComputeTaskChildren(key)
	if err != nil {
		return false, err
	}

	trainChildren := 0

	for _, child := range children {
		if child.Category != asset.ComputeTaskCategory_TASK_TEST && child.Category != asset.ComputeTaskCategory_TASK_PREDICT {
			trainChildren++
		}
		state := newState(&dumbUpdater, child)
		if len(state.AvailableTransitions()) > 0 {
			logger.Debug().Str("childKey", child.Key).Msg("cannot disable model: task has active children")
			return false, nil
		}
	}

	if trainChildren == 0 {
		logger.Debug().Msg("cannot disable model: task has no children")
		return false, nil
	}

	return true, nil
}

func (s *ComputeTaskService) addComputeTaskOutputAsset(output *asset.ComputeTaskOutputAsset) error {
	err := s.GetComputeTaskDBAL().AddComputeTaskOutputAsset(output)
	if err != nil {
		return err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  output.ComputeTaskKey,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK_OUTPUT_ASSET,
		Asset:     &asset.Event_ComputeTaskOutputAsset{ComputeTaskOutputAsset: output},
	}

	return s.GetEventService().RegisterEvents(event)
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

// validateInputs validates that:
// - the task inputs are compatible with the algo inputs
// - the assets referenced by the inputs exist and are of the correct kind
// - the requester has sufficient permissions
func (s *ComputeTaskService) validateInputs(t []*asset.ComputeTaskInput, a map[string]*asset.AlgoInput, owner string, worker string) error {

	seen := make(map[string]bool, 0)

	// Validate inputs
	for _, taskInput := range t {
		identifier := taskInput.Identifier
		algoInput, ok := a[identifier]
		if !ok {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("unknown task input: this identifier was not declared in the Algo: %q", identifier))
		}
		if seen[identifier] && !algoInput.Multiple {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("duplicate task input: this identifier is present multiple times in the task inputs, but it was not declared as \"Multiple\" in the Algo: %q", identifier))
		}

		seen[identifier] = true

		switch algoInput.Kind {
		case asset.AssetKind_ASSET_DATA_SAMPLE:
			// Data samples cannot be validated individually. They will be validated with the data manager, see below.
		case asset.AssetKind_ASSET_DATA_MANAGER:
			if err := s.validateDataManagerInput(taskInput, t, a, owner); err != nil {
				return err
			}
		default:
			if err := s.validateInputRef(algoInput, taskInput, worker); err != nil {
				return err
			}
		}
	}

	// Check there's no required input that was not provided
	for identifier, algoInput := range a {
		if algoInput.Optional {
			continue
		}
		if _, ok := seen[identifier]; !ok {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("missing task input: this identifier is required by the Algo: %q", identifier))
		}
	}

	return nil
}

// validateDataManagerInput validates that the task inputs corresponding to a data manager and data samples are valid, and
// that the requester has sufficient permissions to create the task.
func (s *ComputeTaskService) validateDataManagerInput(dataManagerInput *asset.ComputeTaskInput, inputs []*asset.ComputeTaskInput, a map[string]*asset.AlgoInput, owner string) error {
	dmKey := dataManagerInput.GetAssetKey()
	if dmKey == "" {
		return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: openers must be referenced using an asset key", dataManagerInput.Identifier))
	}

	dsKeys := make([]string, 0)

	// Loop through the data samples
	for _, input := range inputs {
		algoInput, ok := a[input.Identifier]
		if !ok {
			return orcerrors.NewInvalidAsset(fmt.Sprintf(
				"unknown task input: this identifier was not declared in the Algo: %q", input.Identifier))
		}
		if algoInput.Kind != asset.AssetKind_ASSET_DATA_SAMPLE {
			continue
		}
		datasample := input
		dsKey := datasample.GetAssetKey()
		if dsKey == "" {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: data samples must be referenced using an asset key", datasample.Identifier))
		}
		dsKeys = append(dsKeys, datasample.GetAssetKey())
	}

	// Check permissions + check datasamples are compatible with data manager
	datamanager, err := s.getCachedDataManager(dmKey)
	if err != nil {
		return err
	}
	return s.GetDataManagerService().CheckDataManager(datamanager, dsKeys, owner)
}

// validateInputRef validates that the asset referenced by the input exist and is of the correct kind
func (s *ComputeTaskService) validateInputRef(algoInput *asset.AlgoInput, taskInput *asset.ComputeTaskInput, worker string) error {
	var ok bool
	var err error
	identifier := taskInput.Identifier

	switch inputRef := taskInput.Ref.(type) {

	case *asset.ComputeTaskInput_AssetKey:
		switch algoInput.Kind {
		case asset.AssetKind_ASSET_MODEL:
			if _, err = s.GetModelService().GetCheckedModel(inputRef.AssetKey, worker); err != nil {
				return err
			}
		default:
			return orcerrors.NewUnimplemented(fmt.Sprintf("invalid task input %q: invalid asset kind", identifier))
		}

	case *asset.ComputeTaskInput_ParentTaskOutput:
		parentTask, err := s.getCachedComputeTask(inputRef.ParentTaskOutput.ParentTaskKey)
		if err != nil {
			if serr, ok := err.(*orcerrors.OrcError); ok && serr.Kind == orcerrors.ErrNotFound {
				return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: parent task key %v not found: %v", identifier, inputRef.ParentTaskOutput.ParentTaskKey, serr.Error()))
			}
			return err
		}
		var algoOutput *asset.AlgoOutput
		if algoOutput, ok = parentTask.Algo.Outputs[inputRef.ParentTaskOutput.OutputIdentifier]; !ok {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: parent task %v: algo output not found: %q", identifier, inputRef.ParentTaskOutput.ParentTaskKey, inputRef.ParentTaskOutput.OutputIdentifier))
		}
		if algoOutput.Kind != algoInput.Kind {
			return orcerrors.NewInvalidAsset(fmt.Sprintf(
				"invalid task input %q: mismatching task input asset kinds: expecting %v but parent task output has kind %v",
				inputRef.ParentTaskOutput.OutputIdentifier, algoInput.Kind, algoOutput.Kind,
			))
		}
		if algoOutput.Multiple && !algoInput.Multiple {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: multiple task output used as single task input", identifier))
		}
		var o *asset.ComputeTaskOutput
		if o, ok = parentTask.Outputs[inputRef.ParentTaskOutput.OutputIdentifier]; !ok {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: parent task %v: output not found: %q", identifier, inputRef.ParentTaskOutput.ParentTaskKey, inputRef.ParentTaskOutput.OutputIdentifier))
		}
		if !s.GetPermissionService().CanProcess(o.Permissions, worker) {
			return orcerrors.NewPermissionDenied(fmt.Sprintf("invalid task input %q: worker %q doesn't have permission to process output %q of task %v", identifier, worker, inputRef.ParentTaskOutput.ParentTaskKey, inputRef.ParentTaskOutput.OutputIdentifier))
		}

	default:
		return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: an asset or task output reference must be specified", identifier))
	}

	return nil
}

// validateOutputs validates that the task outputs are compatible with the algo outputs
func (s *ComputeTaskService) validateOutputs(
	taskKey string,
	computeTaskOutputs map[string]*asset.ComputeTaskOutput,
	algoOutputs namedAlgoOutputs,
) error {
	seen := make(map[string]bool, len(algoOutputs))

	for identifier, output := range computeTaskOutputs {
		algoOutput, ok := algoOutputs[identifier]
		if !ok {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task %v, unknown task output: this identifier was not declared in the Algo: %q", taskKey, identifier))
		}
		if algoOutput.Kind == asset.AssetKind_ASSET_PERFORMANCE && !output.Permissions.Process.Public {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task %v, invalid task output %q: a PERFORMANCE output should be public", taskKey, identifier))
		}
		if algoOutput.Kind == asset.AssetKind_ASSET_PERFORMANCE && output.Transient {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task %v, invalid task output %q: a PERFORMANCE output cannot be transient", taskKey, identifier))
		}
		seen[identifier] = true
	}

	for identifier := range algoOutputs {
		if _, ok := seen[identifier]; !ok {
			return orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task %v, missing task output: this identifier is required by the Algo: %q", taskKey, identifier))
		}
	}

	return nil
}

// getCachedComputeTask gets the compute task from the cache. If it is not there, it will be fetched from the database and cached.
func (s *ComputeTaskService) getCachedComputeTask(key string) (*asset.ComputeTask, error) {
	if _, ok := s.taskStore[key]; !ok {
		task, err := s.GetComputeTaskDBAL().GetComputeTask(key)
		if err != nil {
			return nil, err
		}
		s.taskStore[key] = task
	}
	return s.taskStore[key], nil
}

// setCompositeData hydrates task specific CompositeTrainTaskData from input
func (s *ComputeTaskService) setCompositeData(taskInput *asset.NewComputeTask, specificInput *asset.NewCompositeTrainTaskData, task *asset.ComputeTask) error {
	datamanager, err := s.getCachedDataManager(specificInput.DataManagerKey)
	if err != nil {
		return err
	}
	err = s.GetDataManagerService().CheckDataManager(datamanager, specificInput.DataSampleKeys, task.Owner)
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

	taskData := &asset.CompositeTrainTaskData{
		DataManagerKey: datamanager.Key,
		DataSampleKeys: specificInput.DataSampleKeys,
	}

	task.Data = &asset.ComputeTask_Composite{
		Composite: taskData,
	}
	task.Worker = datamanager.Owner
	task.LogsPermission = datamanager.LogsPermission

	return nil
}

// setAggregateData hydrates task specific AggregateTrainTaskData from input
func (s *ComputeTaskService) setAggregateData(taskInput *asset.NewComputeTask, input *asset.NewAggregateTrainTaskData, task *asset.ComputeTask, parentTasks []*asset.ComputeTask) error {
	organization, err := s.GetOrganizationService().GetOrganization(input.Worker)
	if err != nil {
		return err
	}

	logsPermission, err := s.GetPermissionService().CreatePermission(task.Owner, &asset.NewPermissions{Public: false})
	if err != nil {
		return err
	}

	for _, p := range parentTasks {
		logsPermission = s.GetPermissionService().UnionPermission(p.LogsPermission, logsPermission)
	}

	taskData := &asset.AggregateTrainTaskData{}
	task.Data = &asset.ComputeTask_Aggregate{
		Aggregate: taskData,
	}
	task.Worker = organization.Id
	task.LogsPermission = logsPermission

	return nil
}

// setTrainData hydrates task specific TrainTaskData from input
func (s *ComputeTaskService) setTrainData(taskInput *asset.NewComputeTask, specificInput *asset.NewTrainTaskData, task *asset.ComputeTask) error {
	datamanager, err := s.getCachedDataManager(specificInput.DataManagerKey)
	if err != nil {
		return err
	}
	err = s.GetDataManagerService().CheckDataManager(datamanager, specificInput.DataSampleKeys, task.Owner)
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

	taskData := &asset.TrainTaskData{
		DataManagerKey: datamanager.Key,
		DataSampleKeys: specificInput.DataSampleKeys,
	}

	task.Data = &asset.ComputeTask_Train{
		Train: taskData,
	}
	task.Worker = datamanager.Owner
	task.LogsPermission = datamanager.LogsPermission

	return nil
}

func (s *ComputeTaskService) setPredictData(taskInput *asset.NewComputeTask, specificInput *asset.NewPredictTaskData, task *asset.ComputeTask) error {
	datamanager, err := s.getCachedDataManager(specificInput.DataManagerKey)
	if err != nil {
		return err
	}
	err = s.GetDataManagerService().CheckDataManager(datamanager, specificInput.DataSampleKeys, task.Owner)
	if err != nil {
		return err
	}

	taskData := &asset.PredictTaskData{
		DataManagerKey: datamanager.Key,
		DataSampleKeys: specificInput.DataSampleKeys,
	}

	task.Data = &asset.ComputeTask_Predict{
		Predict: taskData,
	}
	task.Worker = datamanager.Owner
	task.LogsPermission = datamanager.LogsPermission

	return nil
}

// setTestData hydrates task specific TestTaskData from input
func (s *ComputeTaskService) setTestData(input *asset.NewTestTaskData, task *asset.ComputeTask, parentTasks []*asset.ComputeTask) error {
	datamanager, err := s.getCachedDataManager(input.DataManagerKey)
	if err != nil {
		return err
	}
	err = s.GetDataManagerService().CheckDataManager(datamanager, input.DataSampleKeys, task.Owner)
	if err != nil {
		return err
	}

	taskData := &asset.TestTaskData{
		DataManagerKey: input.DataManagerKey,
		DataSampleKeys: input.DataSampleKeys,
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
	task.ComputePlanKey = parentTasks[0].ComputePlanKey
	// In case of test tasks there is only one parent (see isCompatibleWithParents)
	// and the test task should have the same rank
	task.Rank = parentTasks[0].Rank

	return nil
}

// checkCanProcessParents raises an error if one of the parent is not processable
func (s *ComputeTaskService) checkCanProcessParents(worker string, parentTasks []*asset.ComputeTask, category asset.ComputeTaskCategory) error {
	switch category {
	case asset.ComputeTaskCategory_TASK_AGGREGATE, asset.ComputeTaskCategory_TASK_TEST, asset.ComputeTaskCategory_TASK_TRAIN, asset.ComputeTaskCategory_TASK_PREDICT:
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
			output, ok := p.Outputs[taskIOShared]
			if !ok {
				return orcerrors.NewInternal(fmt.Sprintf("Task %q has no output %s", p.Key, taskIOShared))
			}
			sharedPerms := output.Permissions
			if !s.GetPermissionService().CanProcess(sharedPerms, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process composite parent task %q, worker %q is not allowed to process trunk model by permissions %v", p.Key, worker, sharedPerms,
				))
			}
			output, ok = p.Outputs[taskIOLocal]
			if !ok {
				return orcerrors.NewInternal(fmt.Sprintf("Task %q has no output %s", p.Key, taskIOLocal))
			}
			localPerms := output.Permissions
			if category == asset.ComputeTaskCategory_TASK_TEST && !s.GetPermissionService().CanProcess(localPerms, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process composite parent task %q, worker %q is not allowed to process head model by permissions %v", p.Key, worker, localPerms,
				))
			}
		case *asset.ComputeTask_Aggregate, *asset.ComputeTask_Train:
			output, ok := p.Outputs[taskIOModel]
			if !ok {
				return orcerrors.NewInternal(fmt.Sprintf("Task %q has no output %s", p.Key, taskIOModel))
			}
			permissions := output.Permissions
			if !s.GetPermissionService().CanProcess(permissions, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process parent task %q, worker %q is not allowed to process model by permissions %v", p.Key, worker, permissions,
				))
			}
		case *asset.ComputeTask_Predict:
			output, ok := p.Outputs[taskIOPredictions]
			if !ok {
				return orcerrors.NewInternal(fmt.Sprintf("Task %q has no output %s", p.Key, taskIOModel))
			}
			permissions := output.Permissions
			if !s.GetPermissionService().CanProcess(permissions, worker) {
				return orcerrors.NewPermissionDenied(fmt.Sprintf(
					"cannot process predict parent task %q, worker %q is not allowed to process predictions by permissions %v", p.Key, worker, permissions,
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
				localPerms := p.Outputs[taskIOLocal].Permissions
				if !s.GetPermissionService().CanProcess(localPerms, worker) {
					return orcerrors.NewPermissionDenied(fmt.Sprintf(
						"cannot process composite parent task %q, worker %q is not allowed to process head model by permissions %v", p.Key, worker, localPerms,
					))
				}
			case !hasAggregateParent || i == 1:
				// Without aggregate parent, the second composite parent should contribute the trunk model.
				sharedPerms := p.Outputs[taskIOShared].Permissions
				if !s.GetPermissionService().CanProcess(sharedPerms, worker) {
					return orcerrors.NewPermissionDenied(fmt.Sprintf(
						"cannot process composite parent task %q, worker %q is not allowed to process trunk model by permissions %v", p.Key, worker, sharedPerms,
					))
				}
			}
		case *asset.ComputeTask_Aggregate:
			permissions := p.Outputs[taskIOModel].Permissions
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
	keys := make([]string, len(tasks))

	for i, task := range tasks {
		keys[i] = task.Key
	}

	return s.GetComputeTaskDBAL().GetExistingComputeTaskKeys(keys)
}

// getExistingParentKeys returns the list of parent tasks already persisted.
func (s *ComputeTaskService) getExistingParentKeys(tasks []*asset.NewComputeTask) ([]string, error) {
	parents := []string{}

	for _, task := range tasks {
		parents = append(parents, task.ParentTaskKeys...)
	}

	return s.GetComputeTaskDBAL().GetExistingComputeTaskKeys(parents)
}

// isCompatibleWithParents checks task compatibility with parents tasks
func (s *ComputeTaskService) isCompatibleWithParents(category asset.ComputeTaskCategory, parents []*asset.ComputeTask) bool {
	inputs := map[asset.ComputeTaskCategory]uint32{}

	for _, p := range parents {
		inputs[p.Category]++
	}

	s.GetLogger().Debug().Str("category", category.String()).Interface("parents", inputs).Msg("checking parent compatibility")

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
		return noTest && inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN]+inputs[asset.ComputeTaskCategory_TASK_PREDICT] == 1
	case asset.ComputeTaskCategory_TASK_AGGREGATE:
		return noTest && inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] >= 1
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		return noTest && noTrain && (noParent || compositeOnly || compositeAndAggregate || twoComposites)
	case asset.ComputeTaskCategory_TASK_PREDICT:
		return noTest && inputs[asset.ComputeTaskCategory_TASK_AGGREGATE]+inputs[asset.ComputeTaskCategory_TASK_COMPOSITE]+inputs[asset.ComputeTaskCategory_TASK_TRAIN] == 1
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

func (s *ComputeTaskService) getCachedDataManager(key string) (*asset.DataManager, error) {
	if _, ok := s.dataManagerStore[key]; !ok {
		dm, err := s.GetDataManagerService().GetDataManager(key)
		if err != nil {
			return nil, err
		}
		s.dataManagerStore[key] = dm
	}
	return s.dataManagerStore[key], nil
}

// getInputAsset returns an input asset with the appropriate requested asset kind
func (s *ComputeTaskService) getInputAsset(kind asset.AssetKind, key, identifier string) (*asset.ComputeTaskInputAsset, error) {
	inputAsset := &asset.ComputeTaskInputAsset{
		Identifier: identifier,
	}
	switch kind {
	case asset.AssetKind_ASSET_MODEL:
		model, err := s.GetModelService().GetModel(key)
		if err != nil {
			return nil, err
		}

		inputAsset.Asset = &asset.ComputeTaskInputAsset_Model{Model: model}
		return inputAsset, nil
	case asset.AssetKind_ASSET_DATA_MANAGER:
		manager, err := s.GetDataManagerService().GetDataManager(key)
		if err != nil {
			return nil, err
		}

		inputAsset.Asset = &asset.ComputeTaskInputAsset_DataManager{DataManager: manager}
		return inputAsset, nil
	case asset.AssetKind_ASSET_DATA_SAMPLE:
		sample, err := s.GetDataSampleService().GetDataSample(key)
		if err != nil {
			return nil, err
		}

		inputAsset.Asset = &asset.ComputeTaskInputAsset_DataSample{DataSample: sample}
		return inputAsset, nil
	}

	return nil, orcerrors.NewUnimplemented(fmt.Sprintf("unsupported input kind: %q", kind.String()))
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
	case asset.ComputeTaskCategory_TASK_PREDICT:
		return algoCategory == asset.AlgoCategory_ALGO_PREDICT
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

func (s *ComputeTaskService) getTaskOutputCounter(taskKey string) (persistence.ComputeTaskOutputCounter, error) {
	return s.GetComputeTaskDBAL().CountComputeTaskRegisteredOutputs(taskKey)
}
