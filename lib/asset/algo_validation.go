package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/owkin/orchestrator/lib/errors"
)

// Validate returns an error if the new algo is not valid:
// missing required data, incompatible values, etc.
func (a *NewAlgo) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, is.UUID),
		validation.Field(&a.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&a.Category, validation.In(AlgoCategory_ALGO_SIMPLE, AlgoCategory_ALGO_COMPOSITE, AlgoCategory_ALGO_AGGREGATE, AlgoCategory_ALGO_METRIC, AlgoCategory_ALGO_PREDICT)),
		validation.Field(&a.Description, validation.Required),
		validation.Field(&a.Algorithm, validation.Required),
		validation.Field(&a.Metadata, validation.By(validateMetadata)),
		validation.Field(&a.NewPermissions, validation.Required),
		validation.Field(&a.Inputs, validation.By(validateInputs)),
		validation.Field(&a.Outputs, validation.By(validateOutputs)),
	)
}

func validateInputs(input interface{}) error {
	algoInputs, ok := input.(map[string]*AlgoInput)
	if !ok {
		return errors.NewInvalidAsset("inputs is not a proper map")
	}

	for name, input := range algoInputs {
		err := validation.Validate(name, validation.Required, validation.Length(1, 100))
		if err != nil {
			return err
		}

		err = validation.ValidateStruct(input,
			validation.Field(&input.Kind, validation.In(AssetKind_ASSET_MODEL, AssetKind_ASSET_DATA_SAMPLE, AssetKind_ASSET_DATA_MANAGER)))
		if err != nil {
			return err
		}
	}

	return nil
}

func validateOutputs(input interface{}) error {
	algoOutputs, ok := input.(map[string]*AlgoOutput)
	if !ok {
		return errors.NewInvalidAsset("outputs is not a proper map")
	}

	for name, output := range algoOutputs {
		err := validation.Validate(name, validation.Required, validation.Length(1, 100))
		if err != nil {
			return err
		}

		err = validation.ValidateStruct(output,
			validation.Field(&output.Kind, validation.In(AssetKind_ASSET_MODEL, AssetKind_ASSET_PERFORMANCE)))
		if err != nil {
			return err
		}
	}

	return nil
}
