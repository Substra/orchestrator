package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationValidate(t *testing.T) {
	cases := map[string]struct {
		org   *RegisterOrganizationParam
		valid bool
	}{
		"empty": {&RegisterOrganizationParam{}, true},
		"invalid": {&RegisterOrganizationParam{
			Address: "substra-backend.org-1.com",
		}, false},
		"valid": {&RegisterOrganizationParam{
			Address: "http://substra-backend.org-1.com/",
		}, true},
		"valid_ip": {&RegisterOrganizationParam{
			Address: "http://127.0.0.1:8080",
		}, true},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.org.Validate())
			} else {
				assert.Error(t, c.org.Validate())
			}
		})
	}
}
