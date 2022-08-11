package handlers

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/dbal"
	"github.com/substra/orchestrator/server/standalone/interceptors"
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

	events, paginationToken, err := services.GetEventService().QueryEvents(
		common.NewPagination(params.PageToken, params.PageSize),
		params.Filter,
		params.Sort,
	)
	if err != nil {
		return nil, err
	}

	return &asset.QueryEventsResponse{
		Events:        events,
		NextPageToken: paginationToken,
	}, nil
}

func (s *EventServer) SubscribeToEvents(param *asset.SubscribeToEventsParam, stream asset.EventService_SubscribeToEventsServer) error {
	ctx := stream.Context()

	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return err
	}

	channel, err := commonInterceptors.ExtractChannel(ctx)
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().
		Str("mspid", mspid).
		Str("channel", channel).
		Str("startEventId", param.StartEventId).
		Msg("Subscribing to events")

	// Use a dedicated database connection per SubscribeToEvents request
	// to prevent connection starvation in the pool.
	conn, err := interceptors.ExtractDatabaseConn(ctx)
	if err != nil {
		return err
	}

	d := dbal.New(ctx, nil, conn, channel)
	return d.SubscribeToEvents(param.StartEventId, stream)
}
