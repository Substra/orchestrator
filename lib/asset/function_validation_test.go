package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type functionTestCase struct {
	function  *NewFunction
	valid bool
}

type updateFunctionTestCase struct {
	function  *UpdateFunctionParam
	valid bool
}

func TestFunctionValidate(t *testing.T) {
	validAddressable := &Addressable{
		StorageAddress: "https://somewhere",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	validPerms := &NewPermissions{
		Public:        false,
		AuthorizedIds: []string{"org1"},
	}

	cases := map[string]functionTestCase{
		"emtpy": {&NewFunction{}, false},
		"invalidKey": {&NewFunction{
			Key:            "not36chars",
			Name:           "invalid key",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, false},
		"valid": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, true},
		"invalid_input_kind": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*FunctionInput{
				"test": {Kind: AssetKind_ASSET_COMPUTE_PLAN, Optional: false, Multiple: false},
			},
		}, false},
		"invalid_output_kind": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Outputs: map[string]*FunctionOutput{
				"test": {Kind: AssetKind_ASSET_COMPUTE_PLAN, Multiple: false},
			},
		}, false},
		"invalid_input: data manager + optional": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*FunctionInput{
				"test": {Kind: AssetKind_ASSET_DATA_MANAGER, Optional: true, Multiple: false},
			},
		}, false},
		"invalid_input: data manager + multiple": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*FunctionInput{
				"test": {Kind: AssetKind_ASSET_DATA_MANAGER, Optional: false, Multiple: true},
			},
		}, false},
		"invalid_output: performance + multiple": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Outputs: map[string]*FunctionOutput{
				"test": {Kind: AssetKind_ASSET_PERFORMANCE, Multiple: true},
			},
		}, false},
		"invalid inputs: multiple data managers": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*FunctionInput{
				"datamanager":  {Kind: AssetKind_ASSET_DATA_MANAGER},
				"datamanager2": {Kind: AssetKind_ASSET_DATA_MANAGER},
			},
		}, false},
		"invalid inputs: data manager without data sample": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*FunctionInput{
				"datamanager": {Kind: AssetKind_ASSET_DATA_MANAGER},
			},
		}, false},
		"invalid inputs: data sample without data manager": {&NewFunction{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test function",
			Function:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*FunctionInput{
				"datasamples": {Kind: AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			},
		}, false},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.function.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.function.Validate(), name+" should be invalid")
		}
	}
}

func TestUpdateFunctionValidate(t *testing.T) {
	cases := map[string]updateFunctionTestCase{
		"empty": {&UpdateFunctionParam{}, false},
		"invalidFunctionKey": {&UpdateFunctionParam{
			Key:  "not36chars",
			Name: "Function Name",
		}, false},
		"valid": {&UpdateFunctionParam{
			Key:  "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			Name: "Function Name",
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.function.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.function.Validate(), name+" should be invalid")
		}
	}
}
