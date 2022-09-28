package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type algoTestCase struct {
	algo  *NewAlgo
	valid bool
}

type updateAlgoTestCase struct {
	algo  *UpdateAlgoParam
	valid bool
}

func TestAlgoValidate(t *testing.T) {
	validAddressable := &Addressable{
		StorageAddress: "https://somewhere",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	validPerms := &NewPermissions{
		Public:        false,
		AuthorizedIds: []string{"org1"},
	}

	cases := map[string]algoTestCase{
		"emtpy": {&NewAlgo{}, false},
		"invalidKey": {&NewAlgo{
			Key:            "not36chars",
			Name:           "invalid key",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, false},
		"valid": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, true},
		"invalid_input_kind": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*AlgoInput{
				"test": {Kind: AssetKind_ASSET_COMPUTE_PLAN, Optional: false, Multiple: false},
			},
		}, false},
		"invalid_output_kind": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Outputs: map[string]*AlgoOutput{
				"test": {Kind: AssetKind_ASSET_COMPUTE_PLAN, Multiple: false},
			},
		}, false},
		"invalid_input: data manager + optional": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*AlgoInput{
				"test": {Kind: AssetKind_ASSET_DATA_MANAGER, Optional: true, Multiple: false},
			},
		}, false},
		"invalid_input: data manager + multiple": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*AlgoInput{
				"test": {Kind: AssetKind_ASSET_DATA_MANAGER, Optional: false, Multiple: true},
			},
		}, false},
		"invalid_output: performance + multiple": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Outputs: map[string]*AlgoOutput{
				"test": {Kind: AssetKind_ASSET_PERFORMANCE, Multiple: true},
			},
		}, false},
		"invalid inputs: multiple data managers": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*AlgoInput{
				"datamanager":  {Kind: AssetKind_ASSET_DATA_MANAGER},
				"datamanager2": {Kind: AssetKind_ASSET_DATA_MANAGER},
			},
		}, false},
		"invalid inputs: data manager without data sample": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*AlgoInput{
				"datamanager": {Kind: AssetKind_ASSET_DATA_MANAGER},
			},
		}, false},
		"invalid inputs: data sample without data manager": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*AlgoInput{
				"datasamples": {Kind: AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			},
		}, false},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.algo.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.algo.Validate(), name+" should be invalid")
		}
	}
}

func TestUpdateAlgoValidate(t *testing.T) {
	cases := map[string]updateAlgoTestCase{
		"empty": {&UpdateAlgoParam{}, false},
		"invalidAlgoKey": {&UpdateAlgoParam{
			Key:  "not36chars",
			Name: "Algo Name",
		}, false},
		"valid": {&UpdateAlgoParam{
			Key:  "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			Name: "Algo Name",
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.algo.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.algo.Validate(), name+" should be invalid")
		}
	}
}
