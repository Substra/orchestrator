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

func TestExtractChannel(t *testing.T) {
	ctx := context.TODO()

	ctxWithChannel := context.WithValue(ctx, ctxChannelKey, "mychannel")

	extracted, err := ExtractChannel(ctxWithChannel)
	assert.NoError(t, err, "extraction should not fail")
	assert.Equal(t, "mychannel", extracted, "Channel should be extracted from context")

	_, err = ExtractChannel(ctx)
	assert.Error(t, err, "Extraction should fail on empty context")
}
