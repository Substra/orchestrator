package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

type EventDBAL interface {
	AddEvents(events ...*asset.Event) error
	QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter, sortOrder asset.SortOrder) ([]*asset.Event, common.PaginationToken, error)
}

type EventDBALProvider interface {
	GetEventDBAL() EventDBAL
}
