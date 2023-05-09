package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/service"
)

func TestPerformanceServiceServer(t *testing.T) {
	server := NewPerformanceServer()
	assert.Implements(t, (*asset.PerformanceServiceServer)(nil), server)
}

func TestRegisterPerformance(t *testing.T) {
	ctx, p := getContext()
	ps := new(service.MockPerformanceAPI)

	server := NewPerformanceServer()

	newPerf := &asset.NewPerformance{
		ComputeTaskKey:   "taskUuid",
		MetricKey:        "metricUuid",
		PerformanceValue: 3.14,
	}

	p.On("GetPerformanceService").Return(ps)
	ps.On("RegisterPerformance", newPerf, "requester").Once().Return(&asset.Performance{
		ComputeTaskKey:              newPerf.ComputeTaskKey,
		MetricKey:                   newPerf.MetricKey,
		ComputeTaskOutputIdentifier: newPerf.ComputeTaskOutputIdentifier,
		PerformanceValue:            newPerf.PerformanceValue,
	}, nil)

	_, err := server.RegisterPerformance(ctx, newPerf)
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ps.AssertExpectations(t)
}

func TestGetPerformance(t *testing.T) {
	ctx, p := getContext()
	ps := new(service.MockPerformanceAPI)

	server := NewPerformanceServer()

	perf := &asset.Performance{
		ComputeTaskKey:              "taskUuid",
		MetricKey:                   "metricUuid",
		ComputeTaskOutputIdentifier: "performance",
		PerformanceValue:            3.14,
	}
	param := &asset.QueryPerformancesParam{
		Filter: &asset.PerformanceQueryFilter{
			ComputeTaskKey:              perf.ComputeTaskKey,
			MetricKey:                   perf.MetricKey,
			ComputeTaskOutputIdentifier: perf.ComputeTaskOutputIdentifier,
		},
		PageToken: "",
		PageSize:  100,
	}

	p.On("GetPerformanceService").Return(ps)
	ps.On("QueryPerformances", common.NewPagination(param.PageToken, param.PageSize), param.Filter).Once().Return([]*asset.Performance{}, param.PageToken, nil)

	_, err := server.QueryPerformances(ctx, param)

	assert.NoError(t, err)

	p.AssertExpectations(t)
	ps.AssertExpectations(t)
}
