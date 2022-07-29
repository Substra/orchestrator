package service

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventAPI interface {
	// RegisterEvents allows registering multiple events at once.
	RegisterEvents(...*asset.Event) error
	QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter, sortOrder asset.SortOrder) ([]*asset.Event, common.PaginationToken, error)
}

type EventServiceProvider interface {
	GetEventService() EventAPI
}

type EventDependencyProvider interface {
	persistence.EventDBALProvider
	event.QueueProvider
	TimeServiceProvider
}

type EventService struct {
	EventDependencyProvider
}

func NewEventService(provider EventDependencyProvider) *EventService {
	return &EventService{provider}
}

// RegisterEvents assigns an ID to each event and persist them.
func (s *EventService) RegisterEvents(events ...*asset.Event) error {
	for _, e := range events {
		e.Id = s.GetEventDBAL().NewEventID()
		e.Timestamp = timestamppb.New(s.GetTimeService().GetTransactionTime())
		err := s.GetEventQueue().Enqueue(e)
		if err != nil {
			return err
		}
	}

	err := s.GetEventDBAL().AddEvents(events...)
	if err != nil {
		return err
	}

	return nil
}

func (s *EventService) QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter, sortOrder asset.SortOrder) ([]*asset.Event, common.PaginationToken, error) {
	return s.GetEventDBAL().QueryEvents(p, filter, sortOrder)
}
