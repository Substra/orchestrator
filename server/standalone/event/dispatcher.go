// Package event contains AMQP dispatcher.
package event

import (
	"context"

	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"google.golang.org/protobuf/encoding/protojson"
)

// AMQPDispatcher dispatch events on an AMQP channel
type AMQPDispatcher struct {
	event.Queue
	amqp common.AMQPPublisher
	// channel is the context of assets and computations
	channel    string
	marshaller protojson.MarshalOptions
}

// NewAMQPDispatcher creates a new dispatcher based on given AMQP session.
// channel argument has nothing to do with AMQP but identifies the context of assets and computation events.
func NewAMQPDispatcher(amqp common.AMQPPublisher, channel string) *AMQPDispatcher {
	return &AMQPDispatcher{
		Queue:      new(common.MemoryQueue),
		amqp:       amqp,
		channel:    channel,
		marshaller: protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true},
	}
}

// Dispatch sends events one by one to the AMQP channel
func (d *AMQPDispatcher) Dispatch(ctx context.Context) error {
	logger.Get(ctx).WithField("num_events", d.Len()).WithField("channel", d.channel).Debug("Dispatching events")
	for _, event := range d.GetEvents() {
		// Contextualize the event in a channel
		event.Channel = d.channel

		data, err := d.marshaller.Marshal(event)
		if err != nil {
			return err
		}

		err = d.amqp.Publish(ctx, d.channel, data)
		if err != nil {
			return err
		}
	}

	return nil
}
