package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type profilingTestCase struct {
	function *ProfilingStep
	valid    bool
}

func TestRegisterProfilingStepValidate(t *testing.T) {
	validKey := "08680966-97ae-4573-8b2d-6c4db2b3c532"
	validStep := "build_function"
	validDuration := uint32(1234567)
	cases := map[string]profilingTestCase{
		"emtpy": {&ProfilingStep{}, false},
		"invalidKey": {&ProfilingStep{
			AssetKey: "not36chars",
			Step:     validStep,
			Duration: validDuration,
		}, false},
		"valid": {&ProfilingStep{
			AssetKey: validKey,
			Step:     validStep,
			Duration: validDuration,
		}, true},
		"empty_asset_key": {&ProfilingStep{
			Step:     validStep,
			Duration: validDuration,
		}, false},
		"empty_step": {&ProfilingStep{
			AssetKey: validKey,
			Duration: validDuration,
		}, false},
		"empty_duration": {&ProfilingStep{
			AssetKey: validKey,
			Step:     validStep,
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
