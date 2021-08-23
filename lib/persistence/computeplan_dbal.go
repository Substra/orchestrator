package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

type ComputePlanDBAL interface {
	ComputePlanExists(key string) (bool, error)
	GetComputePlan(key string) (*asset.ComputePlan, error)
	// GetRawComputePlan should return a compute plan without its computed properties (status, task count, etc)
	GetRawComputePlan(key string) (*asset.ComputePlan, error)
	AddComputePlan(plan *asset.ComputePlan) error
	QueryComputePlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error)
}

type ComputePlanDBALProvider interface {
	GetComputePlanDBAL() ComputePlanDBAL
}
