package asset

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/substra/orchestrator/lib/errors"
)

// Validate returns an error if the new function is not valid:
// missing required data, incompatible values, etc.
func (a *NewFunction) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, is.UUID),
		validation.Field(&a.Name, nameValidationRules...),
		validation.Field(&a.Description, validation.Required),
		validation.Field(&a.Function, validation.Required),
		validation.Field(&a.Metadata, validation.By(validateMetadata)),
		validation.Field(&a.NewPermissions, validation.Required),
		validation.Field(&a.Inputs, validation.By(validateInputs)),
		validation.Field(&a.Outputs, validation.By(validateOutputs)),
	)
}

// Validate returns an error if the updated function is not valid:
// missing required data, incompatible values, etc.
func (o *UpdateFunctionParam) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Key, validation.Required, is.UUID),
		validation.Field(&o.Name, nameValidationRules...),
	)
}

func validateInputs(input interface{}) error {
	functionInputs, ok := input.(map[string]*FunctionInput)
	if !ok {
		return errors.NewInvalidAsset("inputs is not a proper map")
	}

	foundDataManager := false
	foundDatasample := false

	for identifier, input := range functionInputs {
		err := validation.Validate(identifier, validation.Required, validation.Length(1, 100))
		if err != nil {
			return err
		}

		err = validation.ValidateStruct(input,
			validation.Field(&input.Kind, validation.In(AssetKind_ASSET_MODEL, AssetKind_ASSET_DATA_SAMPLE, AssetKind_ASSET_DATA_MANAGER)))
		if err != nil {
			return err
		}

		if input.Kind == AssetKind_ASSET_DATA_MANAGER {
			if input.Multiple || input.Optional {
				return errors.NewInvalidAsset(fmt.Sprintf("function input \"%v\" of kind DATA_MANAGER cannot be Multiple or Optional", identifier))
			}
			if foundDataManager {
				return errors.NewInvalidAsset(fmt.Sprintf("cannot have multiple inputs of type %v", AssetKind_ASSET_DATA_MANAGER))
			}
			foundDataManager = true
		}

		if input.Kind == AssetKind_ASSET_DATA_SAMPLE {
			foundDatasample = true
		}
	}

	if foundDataManager != foundDatasample {
		return errors.NewInvalidAsset(fmt.Sprintf("cannot have an input of type %v without an input of type %v, and vice versa", AssetKind_ASSET_DATA_MANAGER, AssetKind_ASSET_DATA_SAMPLE))
	}

	return nil
}

func validateOutputs(input interface{}) error {
	functionOutputs, ok := input.(map[string]*FunctionOutput)
	if !ok {
		return errors.NewInvalidAsset("outputs is not a proper map")
	}

	for identifier, output := range functionOutputs {
		err := validation.Validate(identifier, validation.Required, validation.Length(1, 100))
		if err != nil {
			return err
		}

		err = validation.ValidateStruct(output,
			validation.Field(&output.Kind, validation.In(AssetKind_ASSET_MODEL, AssetKind_ASSET_PERFORMANCE)))
		if err != nil {
			return err
		}

		if output.Kind == AssetKind_ASSET_PERFORMANCE {
			if output.Multiple {
				return errors.NewInvalidAsset(fmt.Sprintf("Function output of kind PERFORMANCE cannot be Multiple. Invalid output: %v", identifier))
			}
		}
	}

	return nil
}
