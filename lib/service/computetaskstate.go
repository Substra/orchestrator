package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/metrics"
)

type taskTransition string

const (
	transitionBuilding            taskTransition = "transitionBuilding"
	transitionWaitingParentTasks  taskTransition = "transitionWaitingParentTasks"
	transitionWaitingExecutorSlot taskTransition = "transitionWaitingExecutorSlot"
	transitionDone                taskTransition = "transitionDone"
	transitionCanceled            taskTransition = "transitionCanceled"
	transitionFailed              taskTransition = "transitionFailed"
	transitionExecuting           taskTransition = "transitionExecuting"
)

// taskStateEvents is the definition of the state machine representing task states
var taskStateEvents = fsm.Events{
	{
		Name: string(transitionBuilding),
		Src:  []string{asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_BUILDING.String(),
	},
	{
		Name: string(transitionWaitingParentTasks),
		Src:  []string{asset.ComputeTaskStatus_STATUS_BUILDING.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS.String(),
	},
	{
		Name: string(transitionCanceled),
		Src:  []string{asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT.String(), asset.ComputeTaskStatus_STATUS_BUILDING.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS.String(), asset.ComputeTaskStatus_STATUS_EXECUTING.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_CANCELED.String(),
	},
	{
		Name: string(transitionWaitingExecutorSlot),
		Src:  []string{asset.ComputeTaskStatus_STATUS_BUILDING.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT.String(),
	},
	{
		Name: string(transitionExecuting),
		Src:  []string{asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_EXECUTING.String(),
	},
	{
		Name: string(transitionDone),
		Src:  []string{asset.ComputeTaskStatus_STATUS_EXECUTING.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_DONE.String(),
	},
	{
		Name: string(transitionFailed),
		Src:  []string{asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT.String(), asset.ComputeTaskStatus_STATUS_BUILDING.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT.String(), asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS.String(), asset.ComputeTaskStatus_STATUS_EXECUTING.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_FAILED.String(),
	},
}

// taskStateUpdater defines a structure capable of handling task updates
type taskStateUpdater interface {
	// On state change will receive the ORIGINAL (before transition) task as first argument
	// and the transition reason as second argument
	// any error should be registered as e.Err
	onStateChange(e *fsm.Event)
	// Recompute children status according to all its parents
	// Task is received as argument
	// any error should be registered as e.Err
	onDone(e *fsm.Event)
	// Set the compute plan to failed when a task fails.
	// Task is received as argument, any error should be registered as e.Err.
	onFailure(e *fsm.Event)
}

// dumbStateUpdater implements taskStateUpdater but does nothing,
// it can be used to evaluate a task state without risking to accidentally update it
type dumbStateUpdater struct{}

func (d *dumbStateUpdater) onStateChange(e *fsm.Event) {}
func (d *dumbStateUpdater) onDone(e *fsm.Event)        {}
func (d *dumbStateUpdater) onFailure(e *fsm.Event)     {}

var dumbUpdater = dumbStateUpdater{}

func newState(updater taskStateUpdater, task *asset.ComputeTask) *fsm.FSM {
	return fsm.NewFSM(
		task.Status.String(),
		taskStateEvents,
		fsm.Callbacks{
			"enter_state":            wrapFsmCallbackContext(updater.onStateChange),
			"after_transitionDone":   wrapFsmCallbackContext(updater.onDone),
			"after_transitionFailed": wrapFsmCallbackContext(updater.onFailure),
		},
	)
}

// wrapFsmCallbackContext wrap our previous updater function with an empty fsm Context (became an argument in v1.0.0)
// We couldn't add this empty parameter in the interface as it would break mock (calling `m.Called(_, e)“)
func wrapFsmCallbackContext(f func(*fsm.Event)) func(context.Context, *fsm.Event) {
	return func(_ context.Context, e *fsm.Event) { f(e) }
}

// ApplyTaskAction apply an asset.ComputeTaskAction to the task.
// It checks the permission and delegate to `applyTaskAction`
// Depending on the current state and action, this may update children tasks
func (s *ComputeTaskService) ApplyTaskAction(key string, action asset.ComputeTaskAction, reason string, requester string) error {
	task, err := s.GetComputeTaskDBAL().GetComputeTask(key)
	if err != nil {
		return err
	}
	if !updateAllowed(task, action, requester) {
		return orcerrors.NewPermissionDenied("only task owner can update it")
	}

	return s.applyTaskAction(task, action, reason)
}

// applyTaskAction apply an asset.ComputeTaskAction to the task.
// This function does NOT check the permissions.
// Depending on the current state and action, this may update children tasks
func (s *ComputeTaskService) applyTaskAction(task *asset.ComputeTask, action asset.ComputeTaskAction, reason string) error {
	var transition taskTransition
	// need to be declared seprately otherwise transition got redeclared
	var err error
	switch action {
	case asset.ComputeTaskAction_TASK_ACTION_CANCELED:
		transition = transitionCanceled
	case asset.ComputeTaskAction_TASK_ACTION_EXECUTING:
		transition = transitionExecuting
	case asset.ComputeTaskAction_TASK_ACTION_FAILED:
		transition = transitionFailed
	case asset.ComputeTaskAction_TASK_ACTION_DONE:
		if task.ComputePlanKey != "" {
			plan, err := s.GetComputePlanService().GetPlan(task.ComputePlanKey)
			if err != nil {
				return err
			}
			if plan.IsTerminated() {
				transition = transitionCanceled
			} else {
				transition = transitionDone
			}
		} else {
			transition = transitionDone
		}
	case asset.ComputeTaskAction_TASK_ACTION_BUILD_STARTED:
		transition = transitionBuilding
	case asset.ComputeTaskAction_TASK_ACTION_BUILD_FINISHED:
		transition, err = s.getTransitionBuildFinished(task.Key)
		if err != nil {
			return err
		}
	default:
		return orcerrors.NewBadRequest("unsupported action")
	}

	if reason == "" {
		reason = "User action"
	}

	return s.applyTaskTransition(task, transition, reason)
}

func (s *ComputeTaskService) getTransitionBuildFinished(taskKey string) (taskTransition, error) {
	parents, err := s.GetComputeTaskDBAL().GetComputeTaskParents(taskKey)

	if err != nil {
		return transitionCanceled, err
	}

	doneParents, err := countParentDone(parents)

	if err != nil {
		return transitionCanceled, err
	}

	if doneParents == len(parents) {
		return transitionWaitingExecutorSlot, nil
	} else {
		return transitionWaitingParentTasks, nil
	}
}

// applyTaskTransition is the internal method allowing any transition (string).
// This allows to use this method with internal only transitions (abort).
func (s *ComputeTaskService) applyTaskTransition(task *asset.ComputeTask, action taskTransition, reason string) error {
	s.GetLogger().Debug().Str("taskKey", task.Key).Str("action", string(action)).Str("reason", reason).Msg("Applying task action")
	state := newState(s, task)
	err := state.Event(context.Background(), string(action), task, reason)

	if err == nil {
		metrics.TaskUpdatedTotal.WithLabelValues(s.GetChannel(), state.Current()).Inc()
	}

	return err
}

// onDone will iterate over task children to update their statuses
func (s *ComputeTaskService) onDone(e *fsm.Event) {
	if len(e.Args) != 2 {
		e.Err = orcerrors.NewInternal(fmt.Sprintf("cannot handle state change with argument: %v", e.Args))
		return
	}
	task, ok := e.Args[0].(*asset.ComputeTask)
	if !ok {
		e.Err = orcerrors.NewInternal("cannot cast argument into task")
		return
	}

	children, err := s.GetComputeTaskDBAL().GetComputeTaskChildren(task.Key)
	if err != nil {
		e.Err = err
		return
	}

	s.GetLogger().Debug().
		Str("taskKey", task.Key).
		Str("taskStatus", task.Status.String()).
		Int("numChildren", len(children)).
		Msg("onDone: updating children statuses")

	for _, child := range children {
		err := s.startChildrenTaskFromParents(task, child)
		if err != nil {
			e.Err = err
			return
		}
	}

	metrics.TaskUpdateCascadeSize.WithLabelValues(s.GetChannel(), string(transitionWaitingExecutorSlot)).Observe(float64(len(children)))
}

// startChildrenTaskFromParents checks which tasks can be started when a parent finishes.
// For each child task, it will check that the function finished building and the other parents statuses are all DONE.
func (s *ComputeTaskService) startChildrenTaskFromParents(triggeringParent, child *asset.ComputeTask) error {
	logger := s.GetLogger().With().
		Str("triggeringParent", triggeringParent.Key).
		Str("triggeringParentStatus", triggeringParent.Status.String()).
		Str("child", child.Key).
		Str("childStatus", child.Status.String()).
		Logger()
	state := newState(s, child)
	if !state.Can(string(transitionWaitingExecutorSlot)) {
		logger.Info().Msg("transitionWaitingExecutorSlot: skipping child due to incompatible state")
		// this is expected as we might go over already failed children (from another parent)
		return nil
	}

	if child.Status != asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS {
		return nil
	}

	err := s.StartDependentTask(child, fmt.Sprintf("Last parent task %s done", triggeringParent.Key))
	if err != nil {
		return err
	}

	return nil
}

func (s *ComputeTaskService) StartDependentTask(child *asset.ComputeTask, reason string) error {
	done, err := s.checkParentTasksDone(child.Key)
	if err != nil {
		return err
	}
	if !done {
		if child.Status != asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS {
			err = s.applyTaskTransition(child, transitionWaitingParentTasks, reason)
			if err != nil {
				return err
			}
		}
		return nil
	}

	logger := s.GetLogger().With().
		Str("child", child.Key).
		Str("childStatus", child.Status.String()).
		Logger()

	state := newState(s, child)
	if !state.Can(string(transitionWaitingExecutorSlot)) {
		logger.Info().Msg("StartDependentTask: skipping child due to incompatible state")
		// this is expected as we might go over already failed children (from another parent)
		return nil
	}

	err = s.applyTaskTransition(child, transitionWaitingExecutorSlot, reason)
	if err != nil {
		return err
	}

	return nil
}

// onStateChange enqueue an orchestration event and saves the task
func (s *ComputeTaskService) onStateChange(e *fsm.Event) {
	if len(e.Args) != 2 {
		e.Err = orcerrors.NewInternal(fmt.Sprintf("cannot handle state change with argument: %v", e.Args))
		return
	}
	task, ok := e.Args[0].(*asset.ComputeTask)
	if !ok {
		e.Err = orcerrors.NewInternal("cannot cast argument into task")
		return
	}
	reason, ok := e.Args[1].(string)
	if !ok {
		e.Err = orcerrors.NewInternal(fmt.Sprintf("cannot cast into string: %v", e.Args[1]))
		return
	}

	statusVal, ok := asset.ComputeTaskStatus_value[e.Dst]
	if !ok {
		// This should not happen since state codes are string representation of statuses
		e.Err = orcerrors.NewInternal(fmt.Sprintf("unknown task status %q", e.Dst))
		return
	}
	task.Status = asset.ComputeTaskStatus(statusVal)

	s.GetLogger().Debug().
		Str("taskKey", task.Key).
		Str("computePlanKey", task.ComputePlanKey).
		Str("newStatus", task.Status.String()).
		Str("taskWorker", task.Worker).
		Str("reason", reason).
		Msg("Updating task status")

	err := s.GetComputeTaskDBAL().UpdateComputeTaskStatus(task.Key, task.Status)
	if err != nil {
		e.Err = err
		return
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  task.Key,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
		Asset:     &asset.Event_ComputeTask{ComputeTask: task},
		Metadata: map[string]string{
			"reason": reason,
		},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		e.Err = err
		return
	}
}

func (s *ComputeTaskService) onFailure(e *fsm.Event) {
	if len(e.Args) != 2 {
		e.Err = orcerrors.NewInternal(fmt.Sprintf("cannot handle state change with argument: %v", e.Args))
		return
	}
	task, ok := e.Args[0].(*asset.ComputeTask)
	if !ok {
		e.Err = orcerrors.NewInternal("cannot cast argument into task")
		return
	}

	err := s.GetComputePlanService().failPlan(task.ComputePlanKey)
	if err != nil {
		orcErr := new(orcerrors.OrcError)
		if errors.As(err, &orcErr) && orcErr.Kind == orcerrors.ErrTerminatedComputePlan {
			s.GetLogger().Debug().
				Str("taskKey", task.Key).
				Str("computePlanKey", task.ComputePlanKey).
				Msg("already terminated compute plan won't be set to failed")

			return
		}

		e.Err = err
	}
}

func countParentDone(parents []*asset.ComputeTask) (int, error) {
	var doneParents = 0

	for _, task := range parents {
		if task.Status == asset.ComputeTaskStatus_STATUS_FAILED || task.Status == asset.ComputeTaskStatus_STATUS_CANCELED {
			return doneParents, orcerrors.NewTerminatedComputeTask(task.Key)
		} else if task.Status == asset.ComputeTaskStatus_STATUS_DONE {
			doneParents++
		}
	}

	return doneParents, nil
}

// getInitialStatus will infer task status from its parents statuses.
func getInitialStatus(parents []*asset.ComputeTask, function *asset.Function) asset.ComputeTaskStatus {
	var status asset.ComputeTaskStatus
	doneParents, err := countParentDone(parents)

	if err != nil {
		return asset.ComputeTaskStatus_STATUS_CANCELED
	}

	switch function.Status {
	case asset.FunctionStatus_FUNCTION_STATUS_READY:
		if doneParents == len(parents) {
			status = asset.ComputeTaskStatus_STATUS_WAITING_FOR_EXECUTOR_SLOT
		} else {
			status = asset.ComputeTaskStatus_STATUS_WAITING_FOR_PARENT_TASKS
		}
	case asset.FunctionStatus_FUNCTION_STATUS_WAITING:
		status = asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT
	case asset.FunctionStatus_FUNCTION_STATUS_BUILDING:
		status = asset.ComputeTaskStatus_STATUS_BUILDING
	case asset.FunctionStatus_FUNCTION_STATUS_FAILED:
		status = asset.ComputeTaskStatus_STATUS_FAILED
	case asset.FunctionStatus_FUNCTION_STATUS_CANCELED:
		status = asset.ComputeTaskStatus_STATUS_CANCELED
	default:
		status = asset.ComputeTaskStatus_STATUS_UNKNOWN
	}

	return status
}

// updateAllowed returns true if the requester can update the task with given action.
// This does not take into account the task status, only ownership.
func updateAllowed(task *asset.ComputeTask, action asset.ComputeTaskAction, requester string) bool {
	switch action {
	case asset.ComputeTaskAction_TASK_ACTION_CANCELED, asset.ComputeTaskAction_TASK_ACTION_FAILED:
		return requester == task.Owner || requester == task.Worker
	case asset.ComputeTaskAction_TASK_ACTION_EXECUTING, asset.ComputeTaskAction_TASK_ACTION_DONE:
		return requester == task.Worker
	default:
		return false
	}
}

func (s *ComputeTaskService) PropagateActionFromFunction(functionKey string, action asset.ComputeTaskAction, reason string, requester string) error {
	//  only select tasks with statuses linked with function status
	var computeTaskStatus asset.ComputeTaskStatus
	if action == asset.ComputeTaskAction_TASK_ACTION_BUILD_STARTED {
		computeTaskStatus = asset.ComputeTaskStatus_STATUS_WAITING_FOR_BUILDER_SLOT
	} else {
		computeTaskStatus = asset.ComputeTaskStatus_STATUS_BUILDING
	}
	tasks, err := s.GetTasksByFunction(functionKey, []asset.ComputeTaskStatus{computeTaskStatus})

	if err != nil {
		return err
	}

	for _, task := range tasks {
		var err error
		if action == asset.ComputeTaskAction_TASK_ACTION_BUILD_FINISHED {
			// Bypass `ApplyTaskAction` as we don't want to run
			err = s.StartDependentTask(task, fmt.Sprintf("Function %s finished building", functionKey))
		} else {
			err = s.applyTaskAction(task, action, reason)
		}
		if err != nil {
			s.GetLogger().Error().
				Err(err).
				Str("functionKey", functionKey).
				Str("taskKey", task.Key).
				Str("action", action.String()).
				Msg("failed to propagate task action when applying function action")
			return err
		}
	}

	return nil
}

func (s *ComputeTaskService) checkParentTasksDone(childKey string) (bool, error) {
	parents, err := s.GetComputeTaskDBAL().GetComputeTaskParents(childKey)
	if err != nil {
		return false, err
	}

	for _, parent := range parents {
		if parent.Status != asset.ComputeTaskStatus_STATUS_DONE {
			s.GetLogger().Debug().
				Str("child", childKey).
				Str("parent", parent.Key).
				Msg("checkParentTasksDone: skipping child due to pending parent")
			// At least one of the parents is not done, so no change for children
			// but no error, this is expected.
			return false, nil
		}
	}

	return true, nil
}
