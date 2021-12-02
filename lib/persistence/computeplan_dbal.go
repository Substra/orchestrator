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
	QueryComputePlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error)
}

type ComputePlanDBALProvider interface {
	GetComputePlanDBAL() ComputePlanDBAL
}

type ComputePlanTaskCount struct {
	Total    int
	Waiting  int
	Todo     int
	Doing    int
	Canceled int
	Failed   int
	Done     int
}

// GetPlanStatus returns the compute plan's status based on its tasks statuses
func (c *ComputePlanTaskCount) GetPlanStatus() asset.ComputePlanStatus {
	if c.Total == 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN
	}

	if c.Done == c.Total {
		return asset.ComputePlanStatus_PLAN_STATUS_DONE
	}

	if c.Failed > 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_FAILED
	}

	if c.Canceled > 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_CANCELED
	}

	if c.Waiting == c.Total {
		return asset.ComputePlanStatus_PLAN_STATUS_WAITING
	}

	if c.Waiting < c.Total && c.Doing == 0 && c.Done == 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_TODO
	}

	return asset.ComputePlanStatus_PLAN_STATUS_DOING
}
