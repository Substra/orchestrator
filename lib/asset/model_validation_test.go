package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelValidate(t *testing.T) {
	validAddressable := &Addressable{
		StorageAddress: "https://somewhere",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	cases := map[string]struct {
		model *NewModel
		valid bool
	}{
		"empty": {&NewModel{}, false},
		"invalid key": {&NewModel{
			Key:                         "not36chars",
			Category:                    ModelCategory_MODEL_SIMPLE,
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "auc",
			Address:                     validAddressable,
		}, false},
		"missing output": {&NewModel{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Category:       ModelCategory_MODEL_SIMPLE,
			ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Address:        validAddressable,
		}, false},
		"valid": {&NewModel{
			Key:                         "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Category:                    ModelCategory_MODEL_SIMPLE,
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "auc",
			Address:                     validAddressable,
		}, true},
		"invalid category": {&NewModel{
			Key:                         "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Category:                    ModelCategory_MODEL_UNKNOWN,
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "auc",
			Address:                     validAddressable,
		}, false},
		"missing category": {&NewModel{
			Key:                         "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "auc",
			Address:                     validAddressable,
		}, true},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.model.Validate())
			} else {
				assert.Error(t, c.model.Validate())
			}
		})
	}
}
