package persistence

import (
	"time"

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
	CancelComputePlan(plan *asset.ComputePlan, ts time.Time) error
}

type ComputePlanDBALProvider interface {
	GetComputePlanDBAL() ComputePlanDBAL
}

type ComputePlanTaskCount struct {
	Total    uint32
	Waiting  uint32
	Todo     uint32
	Doing    uint32
	Canceled uint32
	Failed   uint32
	Done     uint32
}

// GetPlanStatus returns the compute plan's status
func GetPlanStatus(cp *asset.ComputePlan, c *ComputePlanTaskCount) asset.ComputePlanStatus {

	if c.Total == 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_EMPTY
	}

	if c.Done == c.Total {
		return asset.ComputePlanStatus_PLAN_STATUS_DONE
	}

	if c.Failed > 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_FAILED
	}

	if cp.CancelationDate != nil || c.Canceled > 0 {
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
