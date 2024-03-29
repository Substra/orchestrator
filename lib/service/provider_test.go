package service

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/persistence"
)

func newMockedProvider() *MockDependenciesProvider {
	provider := new(MockDependenciesProvider)
	// Unconditionally mock logger
	logger := log.With().Bool("test", true).Logger()
	provider.On("GetLogger").Maybe().Return(&logger)
	// And channel
	provider.On("GetChannel").Maybe().Return("testChannel")

	return provider
}

func TestServiceProviderInit(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	time := new(MockTimeAPI)
	ctx := context.Background()
	ctx = log.With().Bool("test", true).Logger().WithContext(ctx)
	provider := NewProvider(ctx, dbal, time, "testChannel")

	assert.Implements(t, (*OrganizationServiceProvider)(nil), provider, "service provider should provide OrganizationService")
	assert.Implements(t, (*DataSampleServiceProvider)(nil), provider, "service provider should provide DataSampleService")
	assert.Implements(t, (*DataManagerDependencyProvider)(nil), provider, "service provider should provide DataManagerService")
	assert.Implements(t, (*DatasetDependencyProvider)(nil), provider, "service provider should provide DatasetService")
	assert.Implements(t, (*FunctionServiceProvider)(nil), provider, "service provider should provide FunctionService")
	assert.Implements(t, (*ComputeTaskServiceProvider)(nil), provider)
	assert.Implements(t, (*ComputePlanServiceProvider)(nil), provider)
	assert.Implements(t, (*EventServiceProvider)(nil), provider)
}

func TestLazyInstanciation(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	time := new(MockTimeAPI)
	ctx := context.Background()
	ctx = log.With().Bool("test", true).Logger().WithContext(ctx)
	provider := NewProvider(ctx, dbal, time, "testChannel")

	assert.Nil(t, provider.organization, "service should be instanciated when needed")

	s := provider.GetOrganizationService()
	assert.NotNil(t, s, "provider should provide an instance")
	assert.NotNil(t, provider.organization, "provider should reuse the service instance")

	s2 := provider.GetOrganizationService()
	assert.Equal(t, s, s2, "the same instance should be reused")
}
