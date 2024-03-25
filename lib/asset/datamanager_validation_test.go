package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type registerTestCase struct {
	datamanager *NewDataManager
	valid       bool
}

type updateDataManagerTestCase struct {
	dataManager *UpdateDataManagerParam
	valid       bool
}

func TestDataManagerValidate(t *testing.T) {
	validAddressable := &Addressable{
		StorageAddress: "http://somewhere",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	validPermissions := &NewPermissions{
		Public:        false,
		AuthorizedIds: []string{"org1"},
	}

	cases := map[string]registerTestCase{
		"empty": {&NewDataManager{}, false},
		"invalidKey": {&NewDataManager{
			Key:            "invalid key",
			Name:           "Test Data Manager",
			NewPermissions: validPermissions,
			Description:    validAddressable,
			Opener:         validAddressable,
			LogsPermission: validPermissions,
		}, false},
		"valid": {&NewDataManager{
			Key:            "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			Name:           "Test Data Manager",
			NewPermissions: validPermissions,
			Description:    validAddressable,
			Opener:         validAddressable,
			LogsPermission: validPermissions,
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.datamanager.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.datamanager.Validate(), name+" should be invalid")
		}
	}
}

func TestUpdateDataManagerValidate(t *testing.T) {
	cases := map[string]updateDataManagerTestCase{
		"empty": {&UpdateDataManagerParam{}, false},
		"invalidDataManagerKey": {&UpdateDataManagerParam{
			Key:  "not36chars",
			Name: "DataManager Name",
		}, false},
		"valid": {&UpdateDataManagerParam{
			Key:  "834f47c3-2d95-4ccd-a718-7143b64e61c0",
			Name: "DataManager Name",
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.dataManager.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.dataManager.Validate(), name+" should be invalid")
		}
	}
}
