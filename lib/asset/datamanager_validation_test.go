package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type registerTestCase struct {
	datamanager *NewDataManager
	valid       bool
}

type updateTestCase struct {
	datamanager *DataManagerUpdateParam
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
			ObjectiveKey:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			Description:    validAddressable,
			Opener:         validAddressable,
			Type:           "test",
		}, false},
		"invalidObjectiveKey": {&NewDataManager{
			Key:            "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			Name:           "Test Data Manager",
			NewPermissions: validPermissions,
			ObjectiveKey:   "invalid key",
			Description:    validAddressable,
			Opener:         validAddressable,
			Type:           "test",
		}, false},
		"emptyObjectiveKey": {&NewDataManager{
			Key:            "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			Name:           "Test Data Manager",
			NewPermissions: validPermissions,
			Description:    validAddressable,
			Opener:         validAddressable,
			Type:           "test",
		}, true},
		"valid": {&NewDataManager{
			Key:            "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			Name:           "Test Data Manager",
			NewPermissions: validPermissions,
			ObjectiveKey:   "9eef1e88-951a-44fb-944a-c3dbd1d72d85",
			Description:    validAddressable,
			Opener:         validAddressable,
			Type:           "test",
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

func TestDataManagerUpdateValidate(t *testing.T) {
	cases := map[string]updateTestCase{
		"empty": {&DataManagerUpdateParam{}, false},
		"invalidKey": {&DataManagerUpdateParam{
			Key:          "invalid key",
			ObjectiveKey: "9eef1e88-951a-44fb-944a-c3dbd1d72d85",
		}, false},
		"invalidObjectiveKey": {&DataManagerUpdateParam{
			Key:          "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			ObjectiveKey: "invalid key",
		}, false},
		"valid": {&DataManagerUpdateParam{
			Key:          "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			ObjectiveKey: "9eef1e88-951a-44fb-944a-c3dbd1d72d85",
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
