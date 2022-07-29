package distributed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type extractTxIDTestCase struct {
	input    string
	valid    bool
	expected string
}

func TestExtractTxIDFromEventID(t *testing.T) {
	cases := map[string]extractTxIDTestCase{
		"valid": {
			"e00848bc-71c3-422f-b637-cbfc9d2e2042:10df4b99-d09e-4744-9093-6e98b9ec3bdb",
			true,
			"e00848bc-71c3-422f-b637-cbfc9d2e2042",
		},
		"invalid": {
			"foo",
			false,
			"",
		},
		"empty": {
			"",
			false,
			"",
		},
	}

	for name, tc := range cases {
		txID, err := extractTxIDFromEventID(tc.input)

		if tc.valid {
			assert.NoError(t, err, name+" should be valid")
			assert.Equal(t, tc.expected, txID)
		} else {
			assert.Error(t, err, name+" should not be valid")
		}
	}
}
