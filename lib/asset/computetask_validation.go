package asset

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/substra/orchestrator/lib/errors"
)

// Validate returns an error if the NewComputeTask is not valid:
// missing required data, incompatible values, etc.
func (t *NewComputeTask) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.Key, validation.Required, is.UUID),
		validation.Field(&t.FunctionKey, validation.Required, is.UUID),
		validation.Field(&t.ComputePlanKey, validation.Required, is.UUID),
		validation.Field(&t.Metadata, validation.By(validateMetadata)),
		validation.Field(&t.Inputs, validation.By(validateTaskInputs)),
		validation.Field(&t.Outputs, validation.By(validateTaskOutputs)),
	)
}

func (p *ApplyTaskActionParam) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.ComputeTaskKey, validation.Required, is.UUID),
		validation.Field(&p.Action, validation.Required, validation.In(
			// TASK_ACTION_BUILDING, TASK_ACTION_WAITING_FOR_EXECUTION  are managed internally based on function status
			ComputeTaskAction_TASK_ACTION_EXECUTING,
			ComputeTaskAction_TASK_ACTION_FAILED,
			ComputeTaskAction_TASK_ACTION_CANCELED,
			ComputeTaskAction_TASK_ACTION_DONE,
		)),
	)
}

func validateTaskInputs(input interface{}) error {
	inputs, ok := input.([]*ComputeTaskInput)
	if !ok {
		return errors.NewInvalidAsset("inputs is not a proper map")
	}

	for _, input := range inputs {
		err := input.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ComputeTaskInput) Validate() error {
	err := validation.ValidateStruct(i,
		validation.Field(&i.Identifier, validation.Required),
		validation.Field(&i.Ref, validation.Required),
	)
	if err != nil {
		return err
	}

	switch ref := i.Ref.(type) {
	case *ComputeTaskInput_AssetKey:
		return ref.Validate()
	case *ComputeTaskInput_ParentTaskOutput:
		return ref.ParentTaskOutput.Validate()
	default:
		return errors.NewInvalidAsset(fmt.Sprintf("unknown input ref %T", i.Ref))
	}
}

func (i *ComputeTaskInput_AssetKey) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.AssetKey, validation.Required, is.UUID),
	)
}

func (i *ParentTaskOutputRef) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.ParentTaskKey, validation.Required, is.UUID),
		validation.Field(&i.OutputIdentifier, validation.Required),
	)
}

func validateTaskOutputs(input interface{}) error {
	outputs, ok := input.(map[string]*NewComputeTaskOutput)
	if !ok {
		return errors.NewInvalidAsset("outputs is not a proper map")
	}

	for identifier, output := range outputs {
		err := validation.Validate(identifier, validation.Required)
		if err != nil {
			return err
		}

		err = output.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *NewComputeTaskOutput) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Permissions, validation.Required),
	)
}
