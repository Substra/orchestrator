package persistence

import (
	"time"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
)

type ComputePlanDBAL interface {
	ComputePlanExists(key string) (bool, error)
	GetComputePlan(key string) (*asset.ComputePlan, error)
	AddComputePlan(plan *asset.ComputePlan) error
	QueryComputePlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error)
	SetComputePlanName(plan *asset.ComputePlan, name string) error
	CancelComputePlan(plan *asset.ComputePlan, cancelationDate time.Time) error
	FailComputePlan(plan *asset.ComputePlan, failureDate time.Time) error
	IsPlanRunning(key string) (bool, error)
}

type ComputePlanDBALProvider interface {
	GetComputePlanDBAL() ComputePlanDBAL
}
