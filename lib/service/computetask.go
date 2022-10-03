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

// Task statuses in which the inputs are defined
var inputDefinedStatus = []asset.ComputeTaskStatus{
	asset.ComputeTaskStatus_STATUS_DOING,
	asset.ComputeTaskStatus_STATUS_TODO,
	asset.ComputeTaskStatus_STATUS_FAILED,
}

type disabler interface {
	disable(assetKey string) error
}

type namedAlgoOutputs = map[string]*asset.AlgoOutput

// ComputeTaskAPI defines the methods to act on ComputeTasks
type ComputeTaskAPI interface {
	RegisterTasks(tasks []*asset.NewComputeTask, owner string) ([]*asset.ComputeTask, error)
	GetTask(key string) (*asset.ComputeTask, error)
	QueryTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error)
	ApplyTaskAction(key string, action asset.ComputeTaskAction, reason string, requester string) error
	GetInputAssets(key string) ([]*asset.ComputeTaskInputAsset, error)
	DisableOutput(taskKey string, identifier string, requester string) error
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
	orgStore         map[string]*asset.Organization
}

// NewComputeTaskService creates a new service
func NewComputeTaskService(provider ComputeTaskDependencyProvider) *ComputeTaskService {
	return &ComputeTaskService{
		ComputeTaskDependencyProvider: provider,
		algoStore:                     make(map[string]*asset.Algo),
		taskStore:                     make(map[string]*asset.ComputeTask),
		planStore:                     make(map[string]*asset.ComputePlan),
		dataManagerStore:              make(map[string]*asset.DataManager),
		orgStore:                      make(map[string]*asset.Organization),
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
	algo, err := s.GetAlgoService().GetAlgo(task.AlgoKey)
	if err != nil {
		return nil, err
	}

	for _, input := range task.Inputs {
		algoInput, ok := algo.Inputs[input.Identifier]
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

func (s *ComputeTaskService) DisableOutput(taskKey string, identifier string, requester string) error {
	task, err := s.GetTask(taskKey)
	if err != nil {
		return err
	}
	if task.Worker != requester {
		return orcerrors.NewPermissionDenied("only the worker can disable a task output")
	}

	state := newState(&dumbUpdater, task)
	if len(state.AvailableTransitions()) > 0 {
		return orcerrors.NewCannotDisableAsset("cannot disable asset: task not in final state")
	}

	output, found := task.Outputs[identifier]
	if !found {
		return orcerrors.NewCannotDisableAsset(fmt.Sprintf("there is no output identifier %s for task %s", identifier, taskKey))
	}

	if !output.Transient {
		return orcerrors.NewCannotDisableAsset("output is not transient")
	}

	outputAssets, err := s.GetComputeTaskDBAL().GetComputeTaskOutputAssets(taskKey, identifier)
	if err != nil {
		return err
	}

	var service disabler

	// All outputs under the same identifier have the same kind so we use only the first one
	switch outputAssets[0].AssetKind {
	case asset.AssetKind_ASSET_MODEL:
		service = s.GetModelService()
	default:
		// This should not happen since we validate output kinds when creating the task
		return orcerrors.NewCannotDisableAsset(fmt.Sprintf("cannot disable output of kind: %s", outputAssets[0].AssetKind))
	}

	children, err := s.GetComputeTaskDBAL().GetComputeTaskChildren(taskKey)
	if err != nil {
		return err
	}

	if len(children) == 0 {
		return orcerrors.NewCannotDisableAsset("cannot disable output of a task with no children")
	}

	for _, child := range children {
		state := newState(&dumbUpdater, child)
		if len(state.AvailableTransitions()) > 0 {
			return orcerrors.NewCannotDisableAsset("cannot disable asset: child not in final state")
		}
	}

	for _, outputAsset := range outputAssets {
		err := service.disable(outputAsset.AssetKey)
		if err != nil {
			return err
		}
	}
	return nil
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

		for _, parent := range getParentTaskKeys(unsortedTasks[i].Inputs) {
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
			for _, key := range getParentTaskKeys(unsortedTasks[i].Inputs) {
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

	parentTasks, err := s.getRegisteredTasks(getParentTaskKeys(input.Inputs)...)
	if err != nil {
		return nil, err
	}

	status := getInitialStatusFromParents(parentTasks)

	if status == asset.ComputeTaskStatus_STATUS_CANCELED || status == asset.ComputeTaskStatus_STATUS_FAILED {
		return nil, orcerrors.NewError(orcerrors.ErrIncompatibleTaskStatus, fmt.Sprintf("cannot create a task with status %q", status.String()))
	}

	if err := s.allModelsAvailable(parentTasks); err != nil {
		return nil, err
	}

	algo, err := s.getCheckedAlgo(input.AlgoKey, owner)
	if err != nil {
		return nil, err
	}

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

	worker, err := s.getTaskWorker(input, algo)
	if err != nil {
		return nil, err
	}
	// Make sure the organization exists
	_, err = s.getCachedOrganization(worker)
	if err != nil {
		return nil, err
	}

	logsPermissions, err := s.getLogsPermission(owner, parentTasks, input.Inputs, algo.Inputs)
	if err != nil {
		return nil, err
	}

	task := &asset.ComputeTask{
		Key:            input.Key,
		AlgoKey:        algo.Key,
		Category:       input.Category,
		Owner:          owner,
		ComputePlanKey: input.ComputePlanKey,
		Metadata:       input.Metadata,
		Status:         status,
		Rank:           getRank(parentTasks),
		ParentTaskKeys: getParentTaskKeys(input.Inputs),
		CreationDate:   timestamppb.New(s.GetTimeService().GetTransactionTime()),
		Inputs:         input.Inputs,
		Outputs:        outputs,
		Worker:         worker,
		LogsPermission: logsPermissions,
	}

	if err := s.validateInputs(task.Inputs, algo.Inputs, task.Owner, task.Worker); err != nil {
		return nil, err
	}

	if err := s.validateOutputs(task.Key, task.Outputs, algo.Outputs); err != nil {
		return nil, err
	}

	s.taskStore[task.Key] = task

	return task, nil
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
// it will return an error if the algorithm is not processable by the owner.
func (s *ComputeTaskService) getCheckedAlgo(algoKey string, owner string) (*asset.Algo, error) {
	algo, err := s.getCachedAlgo(algoKey)
	if err != nil {
		return nil, err
	}

	canProcess := s.GetPermissionService().CanProcess(algo.Permissions, owner)
	if !canProcess {
		return nil, orcerrors.NewPermissionDenied(fmt.Sprintf("not authorized to process algo %q", algo.Key))
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

		parentTaskAlgo, err := s.getCachedAlgo(parentTask.AlgoKey)
		if err != nil {
			return err
		}

		var algoOutput *asset.AlgoOutput
		if algoOutput, ok = parentTaskAlgo.Outputs[inputRef.ParentTaskOutput.OutputIdentifier]; !ok {
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
		parents = append(parents, getParentTaskKeys(task.Inputs)...)
	}

	return s.GetComputeTaskDBAL().GetExistingComputeTaskKeys(parents)
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

// getCachedAlgo returns a cached version of an algo
// we cache the result to avoid multiple dbal queries on batch registration
func (s *ComputeTaskService) getCachedAlgo(algoKey string) (*asset.Algo, error) {
	if _, ok := s.algoStore[algoKey]; !ok {
		algo, err := s.GetAlgoService().GetAlgo(algoKey)
		if err != nil {
			return nil, err
		}
		s.algoStore[algoKey] = algo
	}
	algo := s.algoStore[algoKey]
	return algo, nil
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

func (s *ComputeTaskService) getCachedOrganization(key string) (*asset.Organization, error) {
	if _, ok := s.orgStore[key]; !ok {
		org, err := s.GetOrganizationService().GetOrganization(key)
		if err != nil {
			return nil, err
		}
		s.orgStore[key] = org
	}
	return s.orgStore[key], nil
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

// getTaskWorker will determine the worker on which the task should execute
func (s *ComputeTaskService) getTaskWorker(input *asset.NewComputeTask, algo *asset.Algo) (string, error) {
	for _, taskInput := range input.Inputs {
		algoInput, ok := algo.Inputs[taskInput.Identifier]
		if !ok {
			return "", orcerrors.NewInvalidAsset(fmt.Sprintf("unknown task input: this identifier was not declared in the Algo: %q", taskInput.Identifier))
		}
		if algoInput.Kind != asset.AssetKind_ASSET_DATA_MANAGER {
			continue
		}

		dm, err := s.getCachedDataManager(taskInput.GetAssetKey())
		if err != nil {
			return "", err
		}
		if input.Worker != "" && input.Worker != dm.Owner {
			return "", orcerrors.NewBadRequest(fmt.Sprintf("Specified worker %q does not match data manager owner: %q", input.Worker, dm.Owner))
		}
		return dm.Owner, nil
	}

	if input.Worker == "" {
		return "", orcerrors.NewBadRequest("Worker cannot be inferred and must be explicitly set")
	}

	return input.Worker, nil
}

// getLogsPermission determines log permission based on datamanager presence.
// If there is a datamanager in inputs, log permission inherit the datamanager's permission.
// If there is no datamanager, log permission is the union of parents log permissions.
func (s *ComputeTaskService) getLogsPermission(owner string, parentTasks []*asset.ComputeTask, taskInputs []*asset.ComputeTaskInput, algoInputs map[string]*asset.AlgoInput) (*asset.Permission, error) {
	// Check for datamanager as input
	for _, taskInput := range taskInputs {
		identifier := taskInput.Identifier
		algoInput, ok := algoInputs[identifier]
		if !ok {
			return nil, orcerrors.NewInvalidAsset(fmt.Sprintf("unknown task input: this identifier was not declared in the Algo: %q", identifier))
		}

		if algoInput.Kind == asset.AssetKind_ASSET_DATA_MANAGER {
			dmKey := taskInput.GetAssetKey()
			if dmKey == "" {
				return nil, orcerrors.NewInvalidAsset(fmt.Sprintf("invalid task input %q: openers must be referenced using an asset key", taskInput.Identifier))
			}
			datamanager, err := s.getCachedDataManager(dmKey)
			if err != nil {
				return nil, err
			}

			return datamanager.LogsPermission, nil
		}
	}

	// Fallback on parent union
	logsPermission, err := s.GetPermissionService().CreatePermission(owner, &asset.NewPermissions{Public: false})
	if err != nil {
		return nil, err
	}
	for _, p := range parentTasks {
		logsPermission = s.GetPermissionService().UnionPermission(p.LogsPermission, logsPermission)
	}

	return logsPermission, nil
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

// getParentTaskKeys returns the parent task keys based on task inputs
func getParentTaskKeys(inputs []*asset.ComputeTaskInput) []string {
	seen := make(map[string]struct{})
	parentKeys := []string{}
	for _, input := range inputs {
		inputRef, ok := input.Ref.(*asset.ComputeTaskInput_ParentTaskOutput)
		if ok {
			parentKey := inputRef.ParentTaskOutput.ParentTaskKey
			if _, ok := seen[parentKey]; !ok {
				parentKeys = append(parentKeys, parentKey)
				seen[parentKey] = struct{}{}
			}
		}
	}

	return parentKeys
}
