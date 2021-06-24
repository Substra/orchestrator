package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
)

type PerformanceDBAL interface {
	AddPerformance(perf *asset.Performance) error
	GetComputeTaskPerformance(key string) (*asset.Performance, error)
}

type PerformanceDBALProvider interface {
	GetPerformanceDBAL() PerformanceDBAL
}
