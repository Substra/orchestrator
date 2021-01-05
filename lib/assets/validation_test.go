// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package assets

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
