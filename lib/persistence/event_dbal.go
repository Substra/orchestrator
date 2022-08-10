package persistence

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
)

type EventDBAL interface {
	NewEventID() string
	AddEvents(events ...*asset.Event) error
	QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter, sortOrder asset.SortOrder) ([]*asset.Event, common.PaginationToken, error)
}

type EventDBALProvider interface {
	GetEventDBAL() EventDBAL
}
