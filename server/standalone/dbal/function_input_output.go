package dbal

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/substra/orchestrator/lib/asset"
)

type sqlFunctionInput struct {
	FunctionKey    string
	Identifier string
	Kind       asset.AssetKind
	Multiple   bool
	Optional   bool
}

func (i *sqlFunctionInput) toFunctionInput() *asset.FunctionInput {
	return &asset.FunctionInput{
		Kind:     i.Kind,
		Multiple: i.Multiple,
		Optional: i.Optional,
	}
}

type sqlFunctionOutput struct {
	FunctionKey    string
	Identifier string
	Kind       asset.AssetKind
	Multiple   bool
}

func (o *sqlFunctionOutput) toFunctionOutput() *asset.FunctionOutput {
	return &asset.FunctionOutput{
		Kind:     o.Kind,
		Multiple: o.Multiple,
	}
}

// insertFunctionInputs insert the function inputs in the database
func (d *DBAL) insertFunctionInputs(functionKey string, inputs map[string]*asset.FunctionInput) error {
	if len(inputs) == 0 {
		return nil
	}

	key, err := uuid.Parse(functionKey)
	if err != nil {
		return err
	}

	rows := make([][]interface{}, 0, len(inputs))

	for identifier, input := range inputs {
		rows = append(rows, []interface{}{key, identifier, input.Kind, input.Multiple, input.Optional})
	}

	_, err = d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"function_inputs"},
		[]string{"function_key", "identifier", "kind", "multiple", "optional"},
		pgx.CopyFromRows(rows),
	)

	return err
}

// insertFunctionOutputs insert the function outputs in the database
func (d *DBAL) insertFunctionOutputs(functionKey string, outputs map[string]*asset.FunctionOutput) error {
	if len(outputs) == 0 {
		return nil
	}

	key, err := uuid.Parse(functionKey)
	if err != nil {
		return err
	}

	rows := make([][]interface{}, 0, len(outputs))

	for identifier, output := range outputs {
		rows = append(rows, []interface{}{key, identifier, output.Kind, output.Multiple})
	}

	_, err = d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"function_outputs"},
		[]string{"function_key", "identifier", "kind", "multiple"},
		pgx.CopyFromRows(rows),
	)

	return err
}

// getFunctionInputs returns the FunctionInputs for the given function keys.
// The returned map has a key for each input function key, e.g.
// {
//   "dcab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &FunctionInput, "datasamples": &FunctionInput },
//   "abcdef01-2345-6789-abcd-ef0123456789": { "model": &FunctionInput, "datasamples": &FunctionInput },
//   "cdab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &FunctionInput, "datasamples": &FunctionInput },
// }

// If an function has no inputs, the corresponding entry in the returned map is an empty map of FunctionInputs.
func (d *DBAL) getFunctionInputs(functionKeys ...string) (map[string]map[string]*asset.FunctionInput, error) {

	res := make(map[string]map[string]*asset.FunctionInput, len(functionKeys))

	for _, functionKey := range functionKeys {
		res[functionKey] = make(map[string]*asset.FunctionInput, 0)
	}

	stmt := getStatementBuilder().
		Select("function_key", "identifier", "kind", "multiple", "optional").
		From("function_inputs").
		Where(sq.Eq{"function_key": functionKeys})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		i := sqlFunctionInput{}
		err = rows.Scan(&i.FunctionKey, &i.Identifier, &i.Kind, &i.Multiple, &i.Optional)
		if err != nil {
			return nil, err
		}
		res[i.FunctionKey][i.Identifier] = i.toFunctionInput()
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// getFunctionOutputs returns the FunctionOutputs for the given function keys.
// The returned map has a key for each output function key, e.g.
//
//	{
//	  "dcab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &FunctionOutput, "othermodel": &FunctionOutput },
//	  "abcdef01-2345-6789-abcd-ef0123456789": { "model": &FunctionOutput, "othermodel": &FunctionOutput },
//	  "cdab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &FunctionOutput, "othermodel": &FunctionOutput },
//	}
//
// If an function has no outputs, the corresponding entry in the returned map is an empty map of FunctionOutputs.
func (d *DBAL) getFunctionOutputs(functionKeys ...string) (map[string]map[string]*asset.FunctionOutput, error) {

	res := make(map[string]map[string]*asset.FunctionOutput, len(functionKeys))

	for _, functionKey := range functionKeys {
		res[functionKey] = make(map[string]*asset.FunctionOutput, 0)
	}

	stmt := getStatementBuilder().
		Select("function_key", "identifier", "kind", "multiple").
		From("function_outputs").
		Where(sq.Eq{"function_key": functionKeys})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		o := sqlFunctionOutput{}
		err = rows.Scan(&o.FunctionKey, &o.Identifier, &o.Kind, &o.Multiple)
		if err != nil {
			return nil, err
		}
		res[o.FunctionKey][o.Identifier] = o.toFunctionOutput()
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// populateFunctionsIO takes multiple function references and decorate them with their inputs and outputs
func (d *DBAL) populateFunctionsIO(functions ...*asset.Function) error {
	keys := make([]string, 0, len(functions))
	for _, function := range functions {
		keys = append(keys, function.Key)
	}

	inputs, err := d.getFunctionInputs(keys...)
	if err != nil {
		return err
	}

	outputs, err := d.getFunctionOutputs(keys...)
	if err != nil {
		return err
	}

	for _, function := range functions {
		function.Inputs = inputs[function.Key]
		function.Outputs = outputs[function.Key]
	}

	return nil
}
