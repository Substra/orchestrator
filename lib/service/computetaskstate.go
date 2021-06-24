package service

import (
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/looplab/fsm"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
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
	// Cascade Canceled status to children. Parent task is given as argument
	// any error should be registered as e.Err
	onCancel(e *fsm.Event)
	// Recompute children status according to all its parents
	// Task is received as argument
	// any error should be registered as e.Err
	onDone(e *fsm.Event)
}

// dumbStateUpdater implements taskStateUpdater but does nothing,
// it can be used to evaluate a task state without risking to accidentaly update it
type dumbStateUpdater struct{}

func (d *dumbStateUpdater) onStateChange(e *fsm.Event) {}
func (d *dumbStateUpdater) onCancel(e *fsm.Event)      {}
func (d *dumbStateUpdater) onDone(e *fsm.Event)        {}

var dumbUpdater = dumbStateUpdater{}

func newState(updater taskStateUpdater, task *asset.ComputeTask) *fsm.FSM {
	return fsm.NewFSM(
		task.Status.String(),
		taskStateEvents,
		fsm.Callbacks{
			"enter_state":              updater.onStateChange,
			"after_transitionCanceled": updater.onCancel,
			"after_transitionFailed":   updater.onCancel,
			"after_transitionDone":     updater.onDone,
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
	default:
		return fmt.Errorf("%w unsupported action", errors.ErrBadRequest)
	}

	if reason == "" {
		reason = "User action"
	}

	task, err := s.GetComputeTaskDBAL().GetComputeTask(key)
	if err != nil {
		return err
	}
	if !updateAllowed(task, action, requester) {
		return fmt.Errorf("only task owner can update it: %w", errors.ErrPermissionDenied)
	}

	return s.applyTaskAction(task, transition, reason)
}

// applyTaskAction is the internal method allowing any transition (string).
// This allows to use this method with internal only transitions (abort).
func (s *ComputeTaskService) applyTaskAction(task *asset.ComputeTask, action taskTransition, reason string) error {
	log.WithField("taskKey", task.Key).WithField("action", action).WithField("reason", reason).Debug("Applying task action")
	state := newState(s, task)
	return state.Event(string(action), task, reason)
}

// onDone will iterate over task children to update their statuses
func (s *ComputeTaskService) onDone(e *fsm.Event) {
	if len(e.Args) != 2 {
		e.Err = fmt.Errorf("cannot handle state change with argument: %v", e.Args)
		return
	}
	task, ok := e.Args[0].(*asset.ComputeTask)
	if !ok {
		e.Err = fmt.Errorf("cannot cast argument into task")
		return
	}

	children, err := s.GetComputeTaskDBAL().GetComputeTaskChildren(task.Key)
	if err != nil {
		e.Err = err
		return
	}

	log.WithFields(
		log.F("taskKey", task.Key),
		log.F("taskStatus", task.Status),
		log.F("numChildren", len(children)),
	).Debug("onDone: updating children statuses")

	for _, child := range children {
		err := s.propagateDone(task, child)
		if err != nil {
			e.Err = err
			return
		}
	}
}

// propagateDone propagates the DONE status of a parent to the task.
// This will iterate over task parents and mark it as TODO if all parents are DONE.
func (s *ComputeTaskService) propagateDone(triggeringParent, child *asset.ComputeTask) error {
	logger := log.WithFields(
		log.F("triggeringParent", triggeringParent.Key),
		log.F("triggeringParentStatus", triggeringParent.Status),
		log.F("child", child.Key),
		log.F("childStatus", child.Status),
	)
	state := newState(s, child)
	if !state.Can(string(transitionTodo)) {
		logger.Info("propagateDone: skipping child due to incompatible state")
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
			logger.WithField("parent", parent.Key).WithField("parentStatus", parent.Status).Debug("propagateDone: skipping child due to pending parent")
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

// onCancel iterate over children to propagate the cancellation
func (s *ComputeTaskService) onCancel(e *fsm.Event) {
	s.cascadeTransition(e, transitionCanceled)
}

func (s *ComputeTaskService) cascadeTransition(e *fsm.Event, transition taskTransition) {
	if len(e.Args) != 2 {
		e.Err = fmt.Errorf("cannot handle state change with argument: %v", e.Args)
		return
	}
	task, ok := e.Args[0].(*asset.ComputeTask)
	if !ok {
		e.Err = fmt.Errorf("cannot cast argument into task")
		return
	}

	children, err := s.GetComputeTaskDBAL().GetComputeTaskChildren(task.Key)
	if err != nil {
		e.Err = err
		return
	}

	log.WithFields(
		log.F("taskKey", task.Key),
		log.F("taskStatus", task.Status),
		log.F("numChildren", len(children)),
		log.F("transition", transition),
	).Debug("Cascading task transition")

	for _, child := range children {
		err := s.applyTaskAction(child, transition, fmt.Sprintf("Cascading status from parent %s", task.Key))
		if err != nil {
			e.Err = err
			return
		}
	}
}

// onStateChange enqueue an orchestration event and saves the task
func (s *ComputeTaskService) onStateChange(e *fsm.Event) {
	if len(e.Args) != 2 {
		e.Err = fmt.Errorf("cannot handle state change with argument: %v", e.Args)
		return
	}
	task, ok := e.Args[0].(*asset.ComputeTask)
	if !ok {
		e.Err = fmt.Errorf("cannot cast into task: %v", e.Args[0])
		return
	}
	reason, ok := e.Args[1].(string)
	if !ok {
		e.Err = fmt.Errorf("cannot cast into string: %v", e.Args[1])
	}

	statusVal, ok := asset.ComputeTaskStatus_value[e.Dst]
	if !ok {
		// This should not happen since state codes are string representation of statuses
		e.Err = fmt.Errorf("unknown task status: %s", e.Dst)
		return
	}
	task.Status = asset.ComputeTaskStatus(statusVal)

	log.WithFields(
		log.F("taskKey", task.Key),
		log.F("newStatus", task.Status),
		log.F("reason", reason),
	).Debug("Updating task status")

	err := s.GetComputeTaskDBAL().UpdateComputeTask(task)
	if err != nil {
		e.Err = err
		return
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  task.Key,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
		Metadata: map[string]string{
			"status": task.Status.String(),
			"reason": reason,
		},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		e.Err = err
		return
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
		return requester == task.Owner
	case asset.ComputeTaskAction_TASK_ACTION_DOING, asset.ComputeTaskAction_TASK_ACTION_FAILED:
		return requester == task.Worker
	default:
		return false
	}
}
