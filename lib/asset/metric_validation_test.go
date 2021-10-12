package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	metric *NewMetric
	valid  bool
}

func TestMetricValidate(t *testing.T) {
	validAddressable := &Addressable{
		StorageAddress: "https://somewhere",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	validPerms := &NewPermissions{
		Public:        false,
		AuthorizedIds: []string{"org1"},
	}

	cases := map[string]testCase{
		"emtpy": {&NewMetric{}, false},
		"invalidKey": {&NewMetric{
			Key:            "not36chars",
			Name:           "invalid key",
			Address:        validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, false},
		"valid": {&NewMetric{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test metric",
			Address:        validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.metric.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.metric.Validate(), name+" should be invalid")
		}
	}
}
