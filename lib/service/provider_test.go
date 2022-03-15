package service

import (
	"testing"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
)

func newMockedProvider() *MockDependenciesProvider {
	provider := new(MockDependenciesProvider)
	// Unconditionally mock logger
	provider.On("GetLogger").Maybe().Return(log.Entry{})
	// And channel
	provider.On("GetChannel").Maybe().Return("testChannel")

	return provider
}

func TestServiceProviderInit(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	queue := new(event.MockQueue)
	time := new(MockTimeAPI)
	provider := NewProvider(log.Entry{}, dbal, queue, time, "testChannel")

	assert.Implements(t, (*NodeServiceProvider)(nil), provider, "service provider should provide NodeService")
	assert.Implements(t, (*MetricServiceProvider)(nil), provider, "service provider should provide MetricService")
	assert.Implements(t, (*DataSampleServiceProvider)(nil), provider, "service provider should provide DataSampleService")
	assert.Implements(t, (*DataManagerDependencyProvider)(nil), provider, "service provider should provide DataManagerService")
	assert.Implements(t, (*DatasetDependencyProvider)(nil), provider, "service provider should provide DatasetService")
	assert.Implements(t, (*AlgoServiceProvider)(nil), provider, "service provider should provide AlgoService")
	assert.Implements(t, (*ComputeTaskServiceProvider)(nil), provider)
	assert.Implements(t, (*ComputePlanServiceProvider)(nil), provider)
	assert.Implements(t, (*EventServiceProvider)(nil), provider)
}

func TestLazyInstanciation(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	queue := new(event.MockQueue)
	time := new(MockTimeAPI)
	provider := NewProvider(log.Entry{}, dbal, queue, time, "testChannel")

	assert.Nil(t, provider.node, "service should be instanciated when needed")

	s := provider.GetNodeService()
	assert.NotNil(t, s, "provider should provide an instance")
	assert.NotNil(t, provider.node, "provider should reuse the service instance")

	s2 := provider.GetNodeService()
	assert.Equal(t, s, s2, "the same instance should be reused")
}
