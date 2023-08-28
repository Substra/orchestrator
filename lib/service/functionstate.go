package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
)

type functionTransition string

const (
	transitionFunctionBuilding functionTransition = "transitionBuilding"
	transitionFunctionReady    functionTransition = "transitionReady"
	transitionFunctionCanceled functionTransition = "transitionCanceled"
	transitionFunctionFailed   functionTransition = "transitionFailed"
)

// functionStateEvents is the definition of the state machine representing function states
var functionStateEvents = fsm.Events{
	{
		Name: string(transitionFunctionCanceled),
		Src:  []string{asset.FunctionStatus_FUNCTION_STATUS_CREATED.String(), asset.FunctionStatus_FUNCTION_STATUS_BUILDING.String()},
		Dst:  asset.FunctionStatus_FUNCTION_STATUS_CANCELED.String(),
	},
	{
		Name: string(transitionFunctionBuilding),
		Src:  []string{asset.FunctionStatus_FUNCTION_STATUS_CREATED.String()},
		Dst:  asset.FunctionStatus_FUNCTION_STATUS_BUILDING.String(),
	},
	{
		Name: string(transitionFunctionReady),
		Src:  []string{asset.FunctionStatus_FUNCTION_STATUS_BUILDING.String()},
		Dst:  asset.FunctionStatus_FUNCTION_STATUS_READY.String(),
	},
	{
		Name: string(transitionFunctionFailed),
		Src:  []string{asset.FunctionStatus_FUNCTION_STATUS_BUILDING.String()},
		Dst:  asset.FunctionStatus_FUNCTION_STATUS_FAILED.String(),
	},
}

// functionStateUpdater defines a structure capable of handling function updates
type functionStateUpdater interface {
	// On state change will receive the ORIGINAL (before transition) function as first argument
	// and the transition reason as second argument
	// any error should be registered as e.Err
	onStateChange(e *fsm.Event)
	// Set the compute plan to failed when a function fails.
	onFailure(e *fsm.Event)
}

func newFunctionState(updater functionStateUpdater, function *asset.Function) *fsm.FSM {
	return fsm.NewFSM(
		function.Status.String(),
		functionStateEvents,
		fsm.Callbacks{
			"enter_state":            wrapFsmCallbackContext(updater.onStateChange),
			"after_transitionFailed": wrapFsmCallbackContext(updater.onFailure),
		},
	)
}

// ApplyFunctionAction apply an asset.FunctioNStatus to the function.
// Depending on the current state and action, this may update children functions
func (s *FunctionService) ApplyFunctionAction(key string, action asset.FunctionAction, reason string, requester string) error {
	var transition functionTransition
	switch action {
	case asset.FunctionAction_FUNCTION_ACTION_BUILDING:
		transition = transitionFunctionBuilding
	case asset.FunctionAction_FUNCTION_ACTION_CANCELED:
		transition = transitionFunctionCanceled
	case asset.FunctionAction_FUNCTION_ACTION_FAILED:
		transition = transitionFunctionFailed
	case asset.FunctionAction_FUNCTION_ACTION_READY:
		transition = transitionFunctionReady
	default:
		return orcerrors.NewBadRequest("unsupported action")
	}

	if reason == "" {
		reason = "User action"
	}

	function, err := s.GetFunctionDBAL().GetFunction(key)
	if err != nil {
		return err
	}
	if requester != function.Owner {
		return orcerrors.NewPermissionDenied("only function owner can update it")
	}

	return s.applyFunctionAction(function, transition, reason)
}

// applyFunctionAction is the internal method allowing any transition (string).
// This allows to use this method with internal only transitions (abort).
func (s *FunctionService) applyFunctionAction(function *asset.Function, action functionTransition, reason string) error {
	s.GetLogger().Debug().Str("functionKey", function.Key).Str("action", string(action)).Str("reason", reason).Msg("Applying function action")
	state := newFunctionState(s, function)
	err := state.Event(context.Background(), string(action), function, reason)

	// if err == nil {
	// 	metrics.FunctionUpdatedTotal.WithLabelValues(s.GetChannel(), state.Current()).Inc()
	// }

	return err
}

// onStateChange enqueue an orchestration event and saves the function
func (s *FunctionService) onStateChange(e *fsm.Event) {
	if len(e.Args) != 2 {
		e.Err = orcerrors.NewInternal(fmt.Sprintf("cannot handle state change with argument: %v", e.Args))
		return
	}
	function, ok := e.Args[0].(*asset.Function)
	if !ok {
		e.Err = orcerrors.NewInternal("cannot cast argument into function")
		return
	}
	reason, ok := e.Args[1].(string)
	if !ok {
		e.Err = orcerrors.NewInternal(fmt.Sprintf("cannot cast into string: %v", e.Args[1]))
		return
	}

	statusVal, ok := asset.FunctionStatus_value[e.Dst]
	if !ok {
		// This should not happen since state codes are string representation of statuses
		e.Err = orcerrors.NewInternal(fmt.Sprintf("unknown function status %q", e.Dst))
		return
	}
	function.Status = asset.FunctionStatus(statusVal)

	s.GetLogger().Debug().
		Str("functionKey", function.Key).
		Str("newStatus", function.Status.String()).
		Str("functionOwner", function.Owner).
		Str("reason", reason).
		Msg("Updating function status")

	err := s.GetFunctionDBAL().UpdateFunctionStatus(function.Key, function.Status)
	if err != nil {
		e.Err = err
		return
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  function.Key,
		AssetKind: asset.AssetKind_ASSET_FUNCTION,
		Asset:     &asset.Event_Function{Function: function},
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

func (s *FunctionService) onFailure(e *fsm.Event) {
	if len(e.Args) != 2 {
		e.Err = orcerrors.NewInternal(fmt.Sprintf("cannot handle state change with argument: %v", e.Args))
		return
	}

	function, ok := e.Args[0].(*asset.Function)
	if !ok {
		e.Err = orcerrors.NewInternal("cannot cast argument into function")
		return
	}

	pagination := common.NewPagination("", 2)
	filter := &asset.TaskQueryFilter{
		Status:      asset.ComputeTaskStatus_STATUS_DOING,
		FunctionKey: function.Key,
	}
	tasks, _, err := s.GetComputeTaskService().QueryTasks(pagination, filter)

	if err != nil {
		orcErr := new(orcerrors.OrcError)
		if errors.As(err, &orcErr) && orcErr.Kind == orcerrors.ErrTerminatedComputePlan {
			s.GetLogger().Debug().
				Str("functionKey", function.Key).
				Msg("cannot get tasks")

			return
		}
	}

	for _, task := range tasks {
		err := s.GetComputeTaskService().ApplyTaskAction(task.Key, asset.ComputeTaskAction_TASK_ACTION_FAILED, "Function building failed", function.Owner)
		if err != nil {
			s.GetLogger().Error().
				Err(err).
				Str("functionKey", function.Key).
				Msg("failed to apply task action when applying function action")
			return
		}
	}

}
