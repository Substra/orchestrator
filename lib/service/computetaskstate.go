package service

import (
	"errors"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/metrics"
)

type taskTransition string

const (
	transitionTodo     taskTransition = "transitionTodo"
	transitionDone     taskTransition = "transitionDone"
	transitionCanceled taskTransition = "transitionCanceled"
	transitionFailed   taskTransition = "transitionFailed"
	transitionDoing    taskTransition = "transitionDoing"
)

// taskStateEvents is the definition of the state machine representing task states
var taskStateEvents = fsm.Events{
	{
		Name: string(transitionCanceled),
		Src:  []string{asset.ComputeTaskStatus_STATUS_TODO.String(), asset.ComputeTaskStatus_STATUS_WAITING.String(), asset.ComputeTaskStatus_STATUS_DOING.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_CANCELED.String(),
	},
	{
		Name: string(transitionTodo),
		Src:  []string{asset.ComputeTaskStatus_STATUS_WAITING.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_TODO.String(),
	},
	{
		Name: string(transitionDoing),
		Src:  []string{asset.ComputeTaskStatus_STATUS_TODO.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_DOING.String(),
	},
	{
		Name: string(transitionDone),
		Src:  []string{asset.ComputeTaskStatus_STATUS_DOING.String()},
		Dst:  asset.ComputeTaskStatus_STATUS_DONE.String(),
	},
	{
		Name: string(transitionFailed),
		Src:  []string{asset.ComputeTaskStatus_STATUS_TODO.String(), asset.ComputeTaskStatus_STATUS_WAITING.String(), asset.ComputeTaskStatus_STATUS_DOING.String()},
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
			"enter_state":            updater.onStateChange,
			"after_transitionDone":   updater.onDone,
			"after_transitionFailed": updater.onFailure,
		},
	)
}

// ApplyTaskAction apply an asset.ComputeTaskAction to the task.
// Depending on the current state and action, this may update children tasks
func (s *ComputeTaskService) ApplyTaskAction(key string, action asset.ComputeTaskAction, reason string, requester string) error {
	var transition taskTransition
	switch action {
	case asset.ComputeTaskAction_TASK_ACTION_CANCELED:
		transition = transitionCanceled
	case asset.ComputeTaskAction_TASK_ACTION_DOING:
		transition = transitionDoing
	case asset.ComputeTaskAction_TASK_ACTION_FAILED:
		transition = transitionFailed
	case asset.ComputeTaskAction_TASK_ACTION_DONE:
		transition = transitionDone
	default:
		return orcerrors.NewBadRequest("unsupported action")
	}

	if reason == "" {
		reason = "User action"
	}

	task, err := s.GetComputeTaskDBAL().GetComputeTask(key)
	if err != nil {
		return err
	}
	if !updateAllowed(task, action, requester) {
		return orcerrors.NewPermissionDenied("only task owner can update it")
	}

	return s.applyTaskAction(task, transition, reason)
}

// applyTaskAction is the internal method allowing any transition (string).
// This allows to use this method with internal only transitions (abort).
func (s *ComputeTaskService) applyTaskAction(task *asset.ComputeTask, action taskTransition, reason string) error {
	s.GetLogger().Debug().Str("taskKey", task.Key).Str("action", string(action)).Str("reason", reason).Msg("Applying task action")
	state := newState(s, task)
	err := state.Event(string(action), task, reason)

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
		err := s.propagateDone(task, child)
		if err != nil {
			e.Err = err
			return
		}
	}

	metrics.TaskUpdateCascadeSize.WithLabelValues(s.GetChannel(), string(transitionTodo)).Observe(float64(len(children)))
}

// propagateDone propagates the DONE status of a parent to the task.
// This will iterate over task parents and mark it as TODO if all parents are DONE.
func (s *ComputeTaskService) propagateDone(triggeringParent, child *asset.ComputeTask) error {
	logger := s.GetLogger().With().
		Str("triggeringParent", triggeringParent.Key).
		Str("triggeringParentStatus", triggeringParent.Status.String()).
		Str("child", child.Key).
		Str("childStatus", child.Status.String()).
		Logger()
	state := newState(s, child)
	if !state.Can(string(transitionTodo)) {
		logger.Info().Msg("propagateDone: skipping child due to incompatible state")
		// this is expected as we might go over already failed children (from another parent)
		return nil
	}

	// loop over parent, only change status if all parents are DONE
	for _, parentKey := range child.ParentTaskKeys {
		if parentKey == triggeringParent.Key {
			// We already know this one is DONE
			continue
		}
		parent, err := s.GetComputeTaskDBAL().GetComputeTask(parentKey)
		if err != nil {
			return err
		}

		if parent.Status != asset.ComputeTaskStatus_STATUS_DONE {
			logger.Debug().
				Str("parent", parent.Key).
				Str("parentStatus", parent.Status.String()).
				Msg("propagateDone: skipping child due to pending parent")
			// At least one of the parents is not done, so no change for children
			// but no error, this is expected.
			return nil
		}
	}
	err := s.applyTaskAction(child, transitionTodo, fmt.Sprintf("Last parent task %s done", triggeringParent.Key))
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

// getInitialStatusFromParents will infer task status from its parents statuses.
func getInitialStatusFromParents(parents []*asset.ComputeTask) asset.ComputeTaskStatus {
	var status asset.ComputeTaskStatus

	statusCount := map[asset.ComputeTaskStatus]int{
		// preset DONE counter to make sure we match TODO status for tasks without parents
		asset.ComputeTaskStatus_STATUS_DONE: 0,
	}

	for _, task := range parents {
		statusCount[task.Status]++
	}

	if c, ok := statusCount[asset.ComputeTaskStatus_STATUS_FAILED]; ok && c > 0 {
		status = asset.ComputeTaskStatus_STATUS_CANCELED
		return status
	}
	if c, ok := statusCount[asset.ComputeTaskStatus_STATUS_CANCELED]; ok && c > 0 {
		status = asset.ComputeTaskStatus_STATUS_CANCELED
		return status
	}

	if c, ok := statusCount[asset.ComputeTaskStatus_STATUS_DONE]; ok && c == len(parents) {
		status = asset.ComputeTaskStatus_STATUS_TODO
	} else {
		status = asset.ComputeTaskStatus_STATUS_WAITING
	}

	return status
}

// updateAllowed returns true if the requester can update the task with given action.
// This does not take into account the task status, only ownership.
func updateAllowed(task *asset.ComputeTask, action asset.ComputeTaskAction, requester string) bool {
	switch action {
	case asset.ComputeTaskAction_TASK_ACTION_CANCELED:
		return requester == task.Owner || requester == task.Worker
	case asset.ComputeTaskAction_TASK_ACTION_DOING, asset.ComputeTaskAction_TASK_ACTION_FAILED, asset.ComputeTaskAction_TASK_ACTION_DONE:
		return requester == task.Worker
	default:
		return false
	}
}
