package dbal

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
)

type sqlTaskInput struct {
	ComputeTaskKey             string
	Identifier                 string
	AssetKey                   string
	ParentTaskKey              string
	ParentTaskOutputIdentifier string
}

func (i *sqlTaskInput) toComputeTaskInput() (*asset.ComputeTaskInput, error) {
	res := &asset.ComputeTaskInput{
		Identifier: i.Identifier,
	}

	if (i.AssetKey != "") == (i.ParentTaskKey != "" && i.ParentTaskOutputIdentifier != "") { // xor
		return nil, errors.NewInternal(fmt.Sprintf("invalid compute task input: either AssetKey or (ParentTaskKey, ParentTaskOutputIdentifier) should be specified, but not both. Row: %v", i))
	}

	if i.AssetKey != "" {
		res.Ref = &asset.ComputeTaskInput_AssetKey{
			AssetKey: i.AssetKey,
		}
	} else {
		res.Ref = &asset.ComputeTaskInput_ParentTaskOutput{
			ParentTaskOutput: &asset.ParentTaskOutputRef{
				ParentTaskKey:    i.ParentTaskKey,
				OutputIdentifier: i.ParentTaskOutputIdentifier,
			},
		}
	}

	return res, nil
}

// insertTaskInputs insert tasks inputs in database in batch mode.
func (d *DBAL) insertTaskInputs(tasks []*asset.ComputeTask) error {
	inputRows, err := getTasksInputRows(tasks)
	if err != nil {
		return err
	}

	_, err = d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"compute_task_inputs"},
		[]string{"compute_task_key", "identifier", "position", "asset_key", "parent_task_key", "parent_task_output_identifier"},
		pgx.CopyFromRows(inputRows),
	)

	return err
}

func getTasksInputRows(tasks []*asset.ComputeTask) ([][]interface{}, error) {

	res := make([][]interface{}, 0)

	for _, task := range tasks {
		taskKey, err := uuid.Parse(task.Key)
		if err != nil {
			return nil, err
		}

		rows, err := getTaskInputRows(taskKey, task.Inputs)
		if err != nil {
			return nil, err
		}

		res = append(res, rows...)
	}

	return res, nil
}

// getTaskInputRows returns task input rows.
func getTaskInputRows(taskKey uuid.UUID, inputs []*asset.ComputeTaskInput) ([][]interface{}, error) {
	res := make([][]interface{}, len(inputs))

	for i, input := range inputs {

		switch ref := input.Ref.(type) {
		case *asset.ComputeTaskInput_AssetKey:
			assetKey, err := uuid.Parse(ref.AssetKey)
			if err != nil {
				return nil, err
			}

			res[i] = []interface{}{
				taskKey,
				input.Identifier,
				i + 1,
				assetKey,
				nil,
				nil,
			}
		case *asset.ComputeTaskInput_ParentTaskOutput:
			parentTaskKey, err := uuid.Parse(ref.ParentTaskOutput.ParentTaskKey)
			if err != nil {
				return nil, err
			}

			res[i] = []interface{}{
				taskKey,
				input.Identifier,
				i + 1,
				nil,
				parentTaskKey,
				ref.ParentTaskOutput.OutputIdentifier,
			}
		default:
			return nil, errors.NewUnimplemented(fmt.Sprintf("invalid compute task input type: %v", input.Ref))
		}
	}

	return res, nil
}

// getTaskInputs returns the ComputeTaskInputs for the given compute task keys.
// The returned map has a key for each input compute task key, e.g.
// {
//   "dcab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": [ &ComputeTaskInput, &ComputeTaskInput ],
//   "abcdef01-2345-6789-abcd-ef0123456789": [ &ComputeTaskInput ],
//   "cdab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": [ &ComputeTaskInput, &ComputeTaskInput, &ComputeTaskInput ],
// }
// If a compute task has no inputs, the corresponding entry in the returned map is an empty list of ComputeTaskInputs.
func (d *DBAL) getTaskInputs(taskKeys ...string) (map[string][]*asset.ComputeTaskInput, error) {

	res := make(map[string][]*asset.ComputeTaskInput, len(taskKeys))

	for _, key := range taskKeys {
		res[key] = make([]*asset.ComputeTaskInput, 0)
	}

	stmt := getStatementBuilder().
		Select(
			"compute_task_key",
			"identifier",
			"COALESCE(asset_key::text, '')",
			"COALESCE(parent_task_key::text, '')",
			"COALESCE(parent_task_output_identifier::text, '')",
		).
		From("compute_task_inputs").
		Where(sq.Eq{"compute_task_key": taskKeys}).
		OrderBy("compute_task_key", "position")

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		i := sqlTaskInput{}
		err = rows.Scan(&i.ComputeTaskKey, &i.Identifier, &i.AssetKey, &i.ParentTaskKey, &i.ParentTaskOutputIdentifier)
		if err != nil {
			return nil, err
		}

		input, err := i.toComputeTaskInput()
		if err != nil {
			return nil, err
		}

		res[i.ComputeTaskKey] = append(res[i.ComputeTaskKey], input)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}