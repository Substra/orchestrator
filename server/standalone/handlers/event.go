package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// EventServer is the gRPC facade to Model manipulation
type EventServer struct {
	asset.UnimplementedEventServiceServer
}

// NewEventServer creates a grpc server
func NewEventServer() *EventServer {
	return &EventServer{}
}

func (s *EventServer) QueryEvents(ctx context.Context, params *asset.QueryEventsParam) (*asset.QueryEventsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	events, paginationToken, err := services.GetEventService().QueryEvents(common.NewPagination(params.PageToken, params.PageSize), params.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryEventsResponse{
		Events:        events,
		NextPageToken: paginationToken,
	}, nil
}
