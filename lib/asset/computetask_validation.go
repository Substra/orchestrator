package asset

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/owkin/orchestrator/lib/errors"
)

// Validate returns an error if the NewComputeTask is not valid:
// missing required data, incompatible values, etc.
func (t *NewComputeTask) Validate() error {
	baseTaskErr := validation.ValidateStruct(t,
		validation.Field(&t.Key, validation.Required, is.UUID),
		validation.Field(&t.Category, validation.Required, validation.In(ComputeTaskCategory_TASK_TRAIN, ComputeTaskCategory_TASK_COMPOSITE, ComputeTaskCategory_TASK_AGGREGATE, ComputeTaskCategory_TASK_TEST)),
		validation.Field(&t.AlgoKey, validation.Required, is.UUID),
		validation.Field(&t.ComputePlanKey, validation.Required, is.UUID),
		validation.Field(&t.Metadata, validation.By(validateMetadata)),
		validation.Field(&t.ParentTaskKeys, validation.Each(is.UUID)),
		validation.Field(&t.Data, validation.Required),
	)

	if baseTaskErr != nil {
		return baseTaskErr
	}

	switch x := t.Data.(type) {
	case *NewComputeTask_Composite:
		return t.Data.(*NewComputeTask_Composite).Composite.Validate()
	case *NewComputeTask_Aggregate:
		return t.Data.(*NewComputeTask_Aggregate).Aggregate.Validate()
	case *NewComputeTask_Test:
		return t.Data.(*NewComputeTask_Test).Test.Validate()
	case *NewComputeTask_Train:
		return t.Data.(*NewComputeTask_Train).Train.Validate()
	default:
		return errors.NewInvalidAsset(fmt.Sprintf("unknown task data %T", x))
	}
}

func (t *NewTestTaskData) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.MetricKeys, validation.Required, validation.Each(validation.Required, is.UUID)),
		validation.Field(&t.DataManagerKey, validation.Required, is.UUID),
		validation.Field(&t.DataSampleKeys, validation.Required, validation.Each(validation.Required, is.UUID)),
	)
}

func (t *NewTrainTaskData) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.DataManagerKey, validation.Required, is.UUID),
		validation.Field(&t.DataSampleKeys, validation.Required, validation.Each(validation.Required, is.UUID)),
	)
}

func (t *NewCompositeTrainTaskData) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.DataManagerKey, validation.Required, is.UUID),
		validation.Field(&t.DataSampleKeys, validation.Required, validation.Each(validation.Required, is.UUID)),
		validation.Field(&t.TrunkPermissions, validation.Required),
	)
}

func (t *NewAggregateTrainTaskData) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.Worker, validation.Required),
	)
}

func (p *ApplyTaskActionParam) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.ComputeTaskKey, validation.Required, is.UUID),
		validation.Field(&p.Action, validation.Required, validation.In(
			ComputeTaskAction_TASK_ACTION_DOING,
			ComputeTaskAction_TASK_ACTION_FAILED,
			ComputeTaskAction_TASK_ACTION_CANCELED,
		)),
	)
}
