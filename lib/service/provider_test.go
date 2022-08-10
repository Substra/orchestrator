package service

import (
	"testing"

	"github.com/go-playground/log/v7"
	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/event"
	"github.com/substra/orchestrator/lib/persistence"
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

	assert.Implements(t, (*OrganizationServiceProvider)(nil), provider, "service provider should provide OrganizationService")
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

	assert.Nil(t, provider.organization, "service should be instanciated when needed")

	s := provider.GetOrganizationService()
	assert.NotNil(t, s, "provider should provide an instance")
	assert.NotNil(t, provider.organization, "provider should reuse the service instance")

	s2 := provider.GetOrganizationService()
	assert.Equal(t, s, s2, "the same instance should be reused")
}
