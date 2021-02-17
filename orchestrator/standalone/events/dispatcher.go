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

package events

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/orchestrator/common"
)

// AMQPDispatcher dispatch events on an AMQP channel
type AMQPDispatcher struct {
	event.Queue
	channel common.AMQPChannel
}

// NewAMQPDispatcher creates a new dispatcher based on given AMQP session
func NewAMQPDispatcher(channel common.AMQPChannel) *AMQPDispatcher {
	return &AMQPDispatcher{
		Queue:   new(common.MemoryQueue),
		channel: channel,
	}
}

// Dispatch sends events one by one to the AMQP channel
func (d *AMQPDispatcher) Dispatch() error {
	log.WithField("num_events", d.Len()).Debug("Dispatching events")
	for _, event := range d.GetEvents() {
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}

		err = d.channel.Publish(data)
		if err != nil {
			return err
		}
	}

	return nil
}
