package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type updateComputePlanTestCase struct {
	computePlan *UpdateComputePlanParam
	valid       bool
}

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

func TestUpdateComputePlanValidate(t *testing.T) {
	cases := map[string]updateComputePlanTestCase{
		"empty": {&UpdateComputePlanParam{}, false},
		"invalidComputePlanKey": {&UpdateComputePlanParam{
			Key:  "not36chars",
			Name: "ComputePlan Name",
		}, false},
		"valid": {&UpdateComputePlanParam{
			Key:  "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			Name: "ComputePlan Name",
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.computePlan.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.computePlan.Validate(), name+" should be invalid")
		}
	}
}
