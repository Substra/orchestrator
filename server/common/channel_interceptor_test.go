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

package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChannelInterceptor(t *testing.T) {
	config := &OrchestratorConfiguration{
		Channels: map[string][]string{
			"mychannel":   {"org1", "org2"},
			"yourchannel": {"org2"},
		},
	}

	interceptor := NewChannelInterceptor(config)

	assert.ElementsMatch(t, interceptor.orgChannels["org1"], []string{"mychannel"})
	assert.ElementsMatch(t, interceptor.orgChannels["org2"], []string{"mychannel", "yourchannel"})
}

func TestCheckOrgBelongsToChannel(t *testing.T) {
	config := &OrchestratorConfiguration{
		Channels: map[string][]string{
			"mychannel":    {"org1", "org2"},
			"yourchannel":  {"org1", "org2"},
			"theirchannel": {"org2", "org3"},
		},
	}
	interceptor := NewChannelInterceptor(config)

	cases := map[string]struct {
		mspid   string
		channel string
		valid   bool
	}{
		"invalid channel": {
			mspid:   "org2",
			channel: "invalid",
			valid:   false,
		},
		"invalid org": {
			mspid:   "org66",
			channel: "mychannel",
			valid:   false,
		},
		"not in channel": {
			mspid:   "org1",
			channel: "theirchannel",
			valid:   false,
		},
		"valid": {
			mspid:   "org1",
			channel: "mychannel",
			valid:   true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := interceptor.checkOrgBelongsToChannel(tc.mspid, tc.channel)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExtractChannel(t *testing.T) {
	ctx := context.TODO()

	ctxWithChannel := context.WithValue(ctx, ctxChannelKey, "mychannel")

	extracted, err := ExtractChannel(ctxWithChannel)
	assert.NoError(t, err, "extraction should not fail")
	assert.Equal(t, "mychannel", extracted, "Channel should be extracted from context")

	_, err = ExtractChannel(ctx)
	assert.Error(t, err, "Extraction should fail on empty context")
}
