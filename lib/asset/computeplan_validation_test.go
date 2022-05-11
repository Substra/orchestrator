package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNewComputePlan(t *testing.T) {
	cases := map[string]struct {
		newComputePlan *NewComputePlan
		valid          bool
	}{
		"empty": {&NewComputePlan{}, false},
		"invalidKey": {&NewComputePlan{
			Key:  "not36chars",
			Name: "The name of my compute plan",
		}, false},
		"valid": {&NewComputePlan{
			Key:  "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name: "The name of my compute plan",
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.newComputePlan.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.newComputePlan.Validate(), name+" should be invalid")
		}
	}
}
