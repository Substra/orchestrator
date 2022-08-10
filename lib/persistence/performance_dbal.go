package persistence

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
)

type PerformanceDBAL interface {
	AddPerformance(perf *asset.Performance, identifier string) error
	QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error)
	PerformanceExists(perf *asset.Performance) (bool, error)
}

type PerformanceDBALProvider interface {
	GetPerformanceDBAL() PerformanceDBAL
}
