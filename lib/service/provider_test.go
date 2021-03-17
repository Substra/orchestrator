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

package service

import (
	"testing"

	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
)

func TestServiceProviderInit(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dispatcher := new(MockDispatcher)
	provider := NewProvider(dbal, dispatcher)

	assert.Implements(t, (*NodeServiceProvider)(nil), provider, "service provider should provide NodeService")
	assert.Implements(t, (*ObjectiveServiceProvider)(nil), provider, "service provider should provide ObjectiveService")
	assert.Implements(t, (*DataSampleServiceProvider)(nil), provider, "service provider should provide DataSampleService")
}

func TestLazyInstanciation(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dispatcher := new(MockDispatcher)
	provider := NewProvider(dbal, dispatcher)

	assert.Nil(t, provider.node, "service should be instanciated when needed")

	s := provider.GetNodeService()
	assert.NotNil(t, s, "provider should provide an instance")
	assert.NotNil(t, provider.node, "provider should reuse the service instance")

	s2 := provider.GetNodeService()
	assert.Equal(t, s, s2, "the same instance should be reused")
}
