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

// Package event contains chaincode related event handling.
// It basically listens chaincode events and convert them into orchestration events.
package event

import (
	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/server/distributed/wallet"
)

// Handler is the signature of the chaincode event callback
type Handler = func(*fab.CCEvent)

// Listener listens to Chaincode events.
type Listener struct {
	contract     *gateway.Contract
	registration fab.Registration
	events       <-chan *fab.CCEvent
	done         chan bool
	onEvent      Handler
}

// NewListener instanciate a Listener listening events on the configured blockchain.
// It filters only events emitted by the chaincode (see ledger.EventName).
// The onEvent callback will be called for every event received.
func NewListener(
	wallet *wallet.Wallet,
	config core.ConfigProvider,
	mspid string,
	channel string,
	chaincode string,
	onEvent Handler,
) (*Listener, error) {
	label := mspid + "-listener"
	wallet.EnsureIdentity(label, mspid)

	gw, err := gateway.Connect(gateway.WithConfig(config), gateway.WithIdentity(wallet, label))

	if err != nil {
		return nil, err
	}

	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		return nil, err
	}

	contract := network.GetContract(chaincode)

	registration, eventStream, err := contract.RegisterEvent(ledger.EventName)
	if err != nil {
		return nil, err
	}

	return &Listener{
		contract:     contract,
		registration: registration,
		events:       eventStream,
		done:         make(chan bool),
		onEvent:      onEvent,
	}, nil
}

// Close will unregister the chaincode listener and properly stop the event listening loop.
func (l *Listener) Close() {
	log.Debug("Closing chaincode event listener")
	l.contract.Unregister(l.registration)
	close(l.done)
}

// Listen will trigger the callback with every event received, until *Listener.Close() is called.
func (l *Listener) Listen() {
	for {
		select {
		case event := <-l.events:
			log.WithField("event", event).Debug("event received")
			l.onEvent(event)
		case <-l.done:
			log.Debug("Stop listening")
			break
		}
	}
}
