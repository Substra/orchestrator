package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	orcledger "github.com/substra/orchestrator/chaincode/ledger"
	forwarder "github.com/substra/orchestrator/forwarder/event"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/server/common/logger"
	"github.com/substra/orchestrator/server/distributed/interceptors"
	"google.golang.org/protobuf/encoding/protojson"
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

func extractTxIDFromEventID(id string) (string, error) {
	s := strings.Split(id, orcledger.EventIDSeparator)

	if len(s) != 2 {
		return "", errors.New("cannot extract TxID from event id string")
	}

	return s[0], nil
}

func queryBlockNumberByTxID(txID string, ccData *forwarder.ListenerChaincodeData) (uint64, error) {
	gw, err := forwarder.ConnectToGateway(ccData)
	if err != nil {
		return 0, err
	}
	defer gw.Close()

	channelProvider := gw.GetChannelProvider(ccData.Channel)
	client, err := ledger.New(channelProvider)
	if err != nil {
		return 0, err
	}

	block, err := client.QueryBlockByTxID(fab.TransactionID(txID))
	if err != nil {
		return 0, err
	}

	return block.Header.Number, nil
}

type EventIndex struct {
	Event forwarder.IndexedEvent
}

func newEventIndex(eventID string, ccData *forwarder.ListenerChaincodeData) (*EventIndex, error) {
	if eventID == "" {
		return &EventIndex{}, nil
	}

	txID, err := extractTxIDFromEventID(eventID)
	if err != nil {
		return nil, err
	}

	blockNum, err := queryBlockNumberByTxID(txID, ccData)
	if err != nil {
		return nil, err
	}

	return &EventIndex{Event: forwarder.IndexedEvent{
		BlockNum:   blockNum,
		TxID:       txID,
		IsIncluded: true,
	}}, nil
}

func (i *EventIndex) GetLastEvent(channel string) forwarder.IndexedEvent {
	return i.Event
}

func (i *EventIndex) SetLastEvent(channel string, event *fab.CCEvent) error {
	return nil
}

func (a *EventAdapter) SubscribeToEvents(param *asset.SubscribeToEventsParam, stream asset.EventService_SubscribeToEventsServer) error {
	ctx := stream.Context()

	ccData, err := interceptors.ExtractChaincodeData(ctx)
	if err != nil {
		return err
	}

	logger.Get(ctx).
		WithField("mspid", ccData.MSPID).
		WithField("chaincode", ccData.Chaincode).
		WithField("channel", ccData.Channel).
		WithField("startEventId", param.StartEventId).
		Info("Subscribing to events")

	eventIdx, err := newEventIndex(param.StartEventId, ccData)
	if err != nil {
		return err
	}

	// As a transaction may have multiple events, make sure we skip events until we reach the last seen
	skipEvent := param.StartEventId != ""

	handle := func(ccEvent *fab.CCEvent) error {
		var rawEvents []json.RawMessage
		err := json.Unmarshal(ccEvent.Payload, &rawEvents)
		if err != nil {
			log.WithError(err).WithField("payload", string(ccEvent.Payload)).Error("Failed to deserialize chaincode event")
			return err
		}

		for _, rawEvent := range rawEvents {
			event := new(asset.Event)
			err = protojson.Unmarshal(rawEvent, event)
			if err != nil {
				log.WithField("rawEvent", string(rawEvent)).WithError(err).Error("failed to deserialize event")
				return err
			}

			skipEvent = skipEvent && param.StartEventId != event.Id
			if skipEvent || param.StartEventId == event.Id {
				log.WithField("event", event).Debug("skipping previously handled event")
				continue
			}

			event.Channel = ccData.Channel

			err = stream.Send(event)
			if err != nil {
				return err
			}
			log.WithField("event", event).Debug("event sent")
		}

		return nil
	}

	listener, err := forwarder.NewListener(ccData, eventIdx, handle)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.WithField("channel", ccData.Channel).WithField("chaincode", ccData.Chaincode).Info("Listening to channel events")
	return listener.Listen(ctx)
}
