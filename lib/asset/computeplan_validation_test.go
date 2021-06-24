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
		"emtpy": {&NewComputePlan{}, false},
		"invalidKey": {&NewComputePlan{
			Key: "not36chars",
		}, false},
		"valid": {&NewComputePlan{
			Key: "08680966-97ae-4573-8b2d-6c4db2b3c532",
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
