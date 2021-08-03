package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// EventAdapter is a grpc server exposing the same Event interface as in standalone,
// but relies on a remote chaincode to actually manage the asset.
type EventAdapter struct {
	asset.UnimplementedEventServiceServer
}

// NewEventAdapter creates a Server
func NewEventAdapter() *EventAdapter {
	return &EventAdapter{}
}

func (a *EventAdapter) QueryEvents(ctx context.Context, query *asset.QueryEventsParam) (*asset.QueryEventsResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.event:QueryEvents"

	response := &asset.QueryEventsResponse{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}
