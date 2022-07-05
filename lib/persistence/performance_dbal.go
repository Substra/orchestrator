package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

type PerformanceDBAL interface {
	AddPerformance(perf *asset.Performance) error
	QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error)
	PerformanceExists(perf *asset.Performance) (bool, error)
}

type PerformanceDBALProvider interface {
	GetPerformanceDBAL() PerformanceDBAL
}
