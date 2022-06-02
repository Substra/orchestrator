package dbal

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
)

type sqlAlgoInput struct {
	AlgoKey    string
	Identifier string
	Kind       asset.AssetKind
	Multiple   bool
	Optional   bool
}

func (i *sqlAlgoInput) toAlgoInput() *asset.AlgoInput {
	return &asset.AlgoInput{
		Kind:     i.Kind,
		Multiple: i.Multiple,
		Optional: i.Optional,
	}
}

type sqlAlgoOutput struct {
	AlgoKey    string
	Identifier string
	Kind       asset.AssetKind
	Multiple   bool
}

func (o *sqlAlgoOutput) toAlgoOutput() *asset.AlgoOutput {
	return &asset.AlgoOutput{
		Kind:     o.Kind,
		Multiple: o.Multiple,
	}
}

// insertAlgoInputs insert the algo inputs in the database
func (d *DBAL) insertAlgoInputs(algoKey string, inputs map[string]*asset.AlgoInput) error {
	if len(inputs) == 0 {
		return nil
	}

	key, err := uuid.Parse(algoKey)
	if err != nil {
		return err
	}

	rows := make([][]interface{}, 0, len(inputs))

	for identifier, input := range inputs {
		rows = append(rows, []interface{}{key, identifier, input.Kind, input.Multiple, input.Optional})
	}

	_, err = d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"algo_inputs"},
		[]string{"algo_key", "identifier", "kind", "multiple", "optional"},
		pgx.CopyFromRows(rows),
	)

	return err
}

// insertAlgoOutputs insert the algo outputs in the database
func (d *DBAL) insertAlgoOutputs(algoKey string, outputs map[string]*asset.AlgoOutput) error {
	if len(outputs) == 0 {
		return nil
	}

	key, err := uuid.Parse(algoKey)
	if err != nil {
		return err
	}

	rows := make([][]interface{}, 0, len(outputs))

	for identifier, output := range outputs {
		rows = append(rows, []interface{}{key, identifier, output.Kind, output.Multiple})
	}

	_, err = d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"algo_outputs"},
		[]string{"algo_key", "identifier", "kind", "multiple"},
		pgx.CopyFromRows(rows),
	)

	return err
}

// getAlgoInputs returns the AlgoInputs for the given algo keys.
// The returned map has a key for each input algo key, e.g.
// {
//   "dcab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &AlgoInput, "datasamples": &AlgoInput },
//   "abcdef01-2345-6789-abcd-ef0123456789": { "model": &AlgoInput, "datasamples": &AlgoInput },
//   "cdab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &AlgoInput, "datasamples": &AlgoInput },
// }

// If an algo has no inputs, the corresponding entry in the returned map is an empty map of AlgoInputs.
func (d *DBAL) getAlgoInputs(algoKeys ...string) (map[string]map[string]*asset.AlgoInput, error) {

	res := make(map[string]map[string]*asset.AlgoInput, len(algoKeys))

	for _, algoKey := range algoKeys {
		res[algoKey] = make(map[string]*asset.AlgoInput, 0)
	}

	stmt := getStatementBuilder().
		Select("algo_key", "identifier", "kind", "multiple", "optional").
		From("algo_inputs").
		Where(sq.Eq{"algo_key": algoKeys})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		i := sqlAlgoInput{}
		err = rows.Scan(&i.AlgoKey, &i.Identifier, &i.Kind, &i.Multiple, &i.Optional)
		if err != nil {
			return nil, err
		}
		res[i.AlgoKey][i.Identifier] = i.toAlgoInput()
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// getAlgoOutputs returns the AlgoOutputs for the given algo keys.
// The returned map has a key for each output algo key, e.g.
// {
//   "dcab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &AlgoOutput, "othermodel": &AlgoOutput },
//   "abcdef01-2345-6789-abcd-ef0123456789": { "model": &AlgoOutput, "othermodel": &AlgoOutput },
//   "cdab4f8f-f8f8-4f8f-8f8f-f8f8f8f8f8f8": { "model": &AlgoOutput, "othermodel": &AlgoOutput },
// }
// If an algo has no outputs, the corresponding entry in the returned map is an empty map of AlgoOutputs.
func (d *DBAL) getAlgoOutputs(algoKeys ...string) (map[string]map[string]*asset.AlgoOutput, error) {

	res := make(map[string]map[string]*asset.AlgoOutput, len(algoKeys))

	for _, algoKey := range algoKeys {
		res[algoKey] = make(map[string]*asset.AlgoOutput, 0)
	}

	stmt := getStatementBuilder().
		Select("algo_key", "identifier", "kind", "multiple").
		From("algo_outputs").
		Where(sq.Eq{"algo_key": algoKeys})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		o := sqlAlgoOutput{}
		err = rows.Scan(&o.AlgoKey, &o.Identifier, &o.Kind, &o.Multiple)
		if err != nil {
			return nil, err
		}
		res[o.AlgoKey][o.Identifier] = o.toAlgoOutput()
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// populateAlgosIO takes multiple algo references and decorate them with their inputs and outputs
func (d *DBAL) populateAlgosIO(algos ...*asset.Algo) error {
	keys := make([]string, 0, len(algos))
	for _, algo := range algos {
		keys = append(keys, algo.Key)
	}

	inputs, err := d.getAlgoInputs(keys...)
	if err != nil {
		return err
	}

	outputs, err := d.getAlgoOutputs(keys...)
	if err != nil {
		return err
	}

	for _, algo := range algos {
		algo.Inputs = inputs[algo.Key]
		algo.Outputs = outputs[algo.Key]
	}

	return nil
}
