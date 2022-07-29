package event

import (
	"context"
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
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
		log.WithError(err).WithField("payload", string(payload)).Error("Failed to deserialize chaincode event")
		return nil
	}

	log.WithField("num_events", len(rawEvents)).Debug("Pushing chaincode events")

	messages := make([][]byte, len(rawEvents))

	for i, rawEvent := range rawEvents {
		event := new(asset.Event)
		err := protojson.Unmarshal(rawEvent, event)
		if err != nil {
			log.WithField("rawEvent", string(rawEvent)).WithError(err).Error("failed to deserialize event")
			continue
		}

		event.Channel = f.channel
		logger := log.WithField("event", event)

		data, err := f.marshaller.Marshal(event)
		if err != nil {
			logger.WithError(err).Error("Failed to serialize")
			continue
		}

		messages[i] = data
		logger.Debug("successfully converted event")
	}
	f.publisher.Publish(context.Background(), f.channel, messages)
	return nil
}
