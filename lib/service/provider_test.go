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
	assert.Implements(t, (*DataManagerDependencyProvider)(nil), provider, "service provider should provide DataManagerService")
	assert.Implements(t, (*DatasetDependencyProvider)(nil), provider, "service provider should provide DatasetService")
	assert.Implements(t, (*AlgoServiceProvider)(nil), provider, "service provider should provide AlgoService")
	assert.Implements(t, (*ComputeTaskServiceProvider)(nil), provider)
	assert.Implements(t, (*ComputePlanServiceProvider)(nil), provider)
	assert.Implements(t, (*EventServiceProvider)(nil), provider)
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
