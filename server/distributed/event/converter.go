// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
)

// Forwarder is responsible for converting chaincode events to AMQP ones and publishing them on an AMQPChannel.
type Forwarder struct {
	channel   string
	publisher common.AMQPPublisher
}

// NewForwarder returns a Converter instance which will publish chaincode events on the given AMQP session.
// It takes a channel which will be used as routing key when publishing messages.
func NewForwarder(channel string, publisher common.AMQPPublisher) *Forwarder {
	return &Forwarder{
		channel:   channel,
		publisher: publisher,
	}
}

// Forward takes a chaincode event, converts it to orchestration events and publish them as AMQP messages.
func (f *Forwarder) Forward(ccEvent *fab.CCEvent) {
	payload := ccEvent.Payload

	events := []*asset.Event{}
	err := json.Unmarshal(payload, &events)

	if err != nil {
		log.WithError(err).Error("Failed to deserialize chaincode event")
	}

	log.WithField("num_events", len(events)).Debug("Pushing chaincode events")

	for _, event := range events {
		logger := log.WithField("event", event)
		event.Channel = f.channel

		data, err := json.Marshal(event)
		if err != nil {
			logger.WithError(err).Error("Failed to serialize")
		}
		err = f.publisher.Publish(f.channel, data)
		if err != nil {
			logger.WithError(err).Error("Failed to push event")
		}
	}
}
