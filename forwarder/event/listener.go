// Package event contains chaincode related event handling.
// It basically listens chaincode events and convert them into orchestration events.
package event

import (
	"context"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/server/distributed/chaincode"
)

// Handler is the signature of the chaincode event callback
type Handler = func(*fab.CCEvent) error

// Listener listens to Chaincode events.
type Listener struct {
	contract     *gateway.Contract
	registration fab.Registration
	events       <-chan *fab.CCEvent
	done         chan bool
	handler      Handler
	channel      string
	eventIdx     Indexer
	logger       log.Entry
}

type ListenerChaincodeData struct {
	Wallet    *chaincode.Wallet
	Config    core.ConfigProvider
	MSPID     string
	Channel   string
	Chaincode string
}

// ConnectToGateway connects to the Fabric gateway using the provided ListenerChaincodeData
func ConnectToGateway(ccData *ListenerChaincodeData, options ...gateway.Option) (*gateway.Gateway, error) {
	label := ccData.MSPID + "-listener"
	err := ccData.Wallet.EnsureIdentity(label, ccData.MSPID)
	if err != nil {
		return nil, err
	}

	return gateway.Connect(
		gateway.WithConfig(ccData.Config),
		gateway.WithIdentity(ccData.Wallet, label),
		options...,
	)
}

// NewListener instanciate a Listener listening events on the configured blockchain.
// It filters only events emitted by the chaincode (see ledger.EventName).
// The onEvent callback will be called for every event received.
func NewListener(
	ccData *ListenerChaincodeData,
	eventIdx Indexer,
	handler Handler,
) (*Listener, error) {
	logger := log.WithFields(
		log.F("channel", ccData.Channel),
		log.F("chaincode", ccData.Chaincode),
	)

	// For new index (without referenced events), this will default to block 0
	checkpoint := eventIdx.GetLastEvent(ccData.Channel)

	logger.WithField("lastEventBlock", checkpoint.BlockNum).WithField("lastTxID", checkpoint.TxID).Debug("instanciating event listener")

	gw, err := ConnectToGateway(ccData, gateway.WithBlockNum(checkpoint.BlockNum))
	if err != nil {
		return nil, err
	}
	defer gw.Close()

	network, err := gw.GetNetwork(ccData.Channel)
	if err != nil {
		return nil, err
	}

	contract := network.GetContract(ccData.Chaincode)

	registration, eventStream, err := contract.RegisterEvent(ledger.EventName)
	if err != nil {
		return nil, err
	}

	return &Listener{
		contract:     contract,
		registration: registration,
		events:       eventStream,
		done:         make(chan bool),
		handler:      handler,
		channel:      ccData.Channel,
		logger:       logger,
		eventIdx:     eventIdx,
	}, nil
}

// Close will unregister the chaincode listener and properly stop the event listening loop.
func (l *Listener) Close() {
	l.logger.Debug("Closing chaincode event listener")
	l.contract.Unregister(l.registration)
	close(l.done)
}

// Listen will trigger the callback with every event received, until *Listener.Close() is called.
func (l *Listener) Listen(ctx context.Context) error {
	checkpoint := l.eventIdx.GetLastEvent(l.channel)
	// As a block may have multiple transactions, make sure we skip events until we reach the last seen
	skipEvent := checkpoint.TxID != ""
	for {
		select {
		case event := <-l.events:
			logger := l.logger.WithFields(
				log.F("eventName", event.EventName),
				log.F("blockNumber", event.BlockNumber),
				log.F("source", event.SourceURL),
				log.F("txID", event.TxID),
			)
			skipEvent = skipEvent && event.TxID != checkpoint.TxID
			if skipEvent || (event.TxID == checkpoint.TxID && !checkpoint.IsIncluded) {
				logger.Debug("skipping previously handled event")
				break
			}

			logger.Debug("handling event")

			err := l.handler(event)
			if err != nil {
				return err
			}

			err = l.eventIdx.SetLastEvent(l.channel, event)
			if err != nil {
				logger.WithError(err).Error("cannot track event")
			}
		case <-l.done:
			l.logger.Debug("Listener done: stop listening")
			return nil
		case <-ctx.Done():
			l.logger.Debug("Context done: stop listening")
			return ctx.Err()
		}
	}
}
