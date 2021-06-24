package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

func TestPerformanceServiceServer(t *testing.T) {
	server := NewPerformanceServer()
	assert.Implements(t, (*asset.PerformanceServiceServer)(nil), server)
}

func TestRegisterPerformance(t *testing.T) {
	ctx, p := getContext()
	ps := new(service.MockPerformanceService)

	server := NewPerformanceServer()

	newPerf := &asset.NewPerformance{ComputeTaskKey: "uuid", PerformanceValue: 3.14}

	p.On("GetPerformanceService").Return(ps)
	ps.On("RegisterPerformance", newPerf, "requester").Once().Return(&asset.Performance{ComputeTaskKey: "uuid"}, nil)

	_, err := server.RegisterPerformance(ctx, newPerf)
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ps.AssertExpectations(t)
}

func TestGetPerformance(t *testing.T) {
	ctx, p := getContext()
	ps := new(service.MockPerformanceService)

	server := NewPerformanceServer()

	perf := &asset.Performance{ComputeTaskKey: "uuid", PerformanceValue: 3.14}

	p.On("GetPerformanceService").Return(ps)
	ps.On("GetComputeTaskPerformance", "uuid").Once().Return(perf, nil)

	_, err := server.GetComputeTaskPerformance(ctx, &asset.GetComputeTaskPerformanceParam{ComputeTaskKey: "uuid"})
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ps.AssertExpectations(t)
}
