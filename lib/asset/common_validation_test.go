package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressableValidation(t *testing.T) {
	emptyAddressable := &Addressable{}
	invalidChecksum := &Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "1234",
	}
	invalidStorage := &Addressable{
		StorageAddress: "0698796898",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	validAddressable := &Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	assert.Error(t, emptyAddressable.Validate(), "empty object is invalid")
	assert.Error(t, invalidChecksum.Validate(), "checksum should be valid")
	assert.Error(t, invalidStorage.Validate(), "storage address should be valid")
	assert.NoError(t, validAddressable.Validate(), "validation should pass")
}

func TestPermissionsValidation(t *testing.T) {
	emptyPermissions := &Permissions{}
	complete := &Permissions{Process: &Permission{Public: true, AuthorizedIds: []string{}}}

	assert.Error(t, emptyPermissions.Validate(), "empty object is invalid")
	assert.NoError(t, complete.Validate())
}
