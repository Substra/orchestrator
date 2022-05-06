package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type algoTestCase struct {
	algo  *NewAlgo
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
			Category:       AlgoCategory_ALGO_SIMPLE,
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, false},
		"invalidCategory": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "invalid category",
			Category:       23,
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, false},
		"valid": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Category:       AlgoCategory_ALGO_AGGREGATE,
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, true},
		"invalid_input_kind": {&NewAlgo{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test algo",
			Category:       AlgoCategory_ALGO_SIMPLE,
			Algorithm:      validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
			Inputs: map[string]*AlgoInput{
				"test": {Kind: AssetKind_ASSET_COMPUTE_PLAN, Optional: false, Multiple: false},
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
