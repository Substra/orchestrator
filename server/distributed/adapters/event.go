package adapters

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/server/distributed/chaincode"
	"github.com/substra/orchestrator/server/distributed/interceptors"
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
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.event:QueryEvents"

	response := &asset.QueryEventsResponse{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

func (a *EventAdapter) SubscribeToEvents(param *asset.SubscribeToEventsParam, stream asset.EventService_SubscribeToEventsServer) error {
	ctx := stream.Context()

	ccData, err := interceptors.ExtractChaincodeData(ctx)
	if err != nil {
		return err
	}

	logger := log.Ctx(ctx).
		With().
		Str("mspid", ccData.MSPID).
		Str("chaincode", ccData.Chaincode).
		Str("channel", ccData.Channel).
		Str("startEventId", param.StartEventId).
		Logger()

	logger.Info().Msg("subscribing to events")

	listener, err := chaincode.NewListener(ccData, param.StartEventId, stream.Send)
	if err != nil {
		return err
	}
	defer listener.Close()

	logger.Info().Msg("listening to channel events")
	return listener.Listen(ctx)
}
