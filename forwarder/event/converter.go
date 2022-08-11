package event

import (
	"context"
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/server/common"
	"google.golang.org/protobuf/encoding/protojson"
)

// Forwarder is responsible for converting chaincode events to AMQP ones and publishing them on an AMQPChannel.
type Forwarder struct {
	channel    string
	publisher  common.AMQPPublisher
	marshaller protojson.MarshalOptions
}

// NewForwarder returns a Converter instance which will publish chaincode events on the given AMQP session.
// It takes a channel which will be used as routing key when publishing messages.
func NewForwarder(channel string, publisher common.AMQPPublisher) *Forwarder {
	return &Forwarder{
		channel:    channel,
		publisher:  publisher,
		marshaller: protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true},
	}
}

// Forward takes a chaincode event, converts it to orchestration events and publish them as AMQP messages.
func (f *Forwarder) Forward(ccEvent *fab.CCEvent) error {
	payload := ccEvent.Payload

	rawEvents := []json.RawMessage{}
	err := json.Unmarshal(payload, &rawEvents)

	if err != nil {
		log.Error().Err(err).Str("payload", string(payload)).Msg("Failed to deserialize chaincode event")
		return nil
	}

	log.Debug().Int("numEvents", len(rawEvents)).Msg("Pushing chaincode events")

	messages := make([][]byte, len(rawEvents))

	for i, rawEvent := range rawEvents {
		event := new(asset.Event)
		err := protojson.Unmarshal(rawEvent, event)
		if err != nil {
			log.Error().Str("rawEvent", string(rawEvent)).Err(err).Msg("failed to deserialize event")
			continue
		}

		event.Channel = f.channel
		logger := log.With().Interface("event", event).Logger()

		data, err := f.marshaller.Marshal(event)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to serialize")
			continue
		}

		messages[i] = data
		logger.Debug().Msg("successfully converted event")
	}
	f.publisher.Publish(context.Background(), f.channel, messages)
	return nil
}
