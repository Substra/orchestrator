package chaincode

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	orcledger "github.com/substra/orchestrator/chaincode/ledger"
	"github.com/substra/orchestrator/lib/asset"
	"google.golang.org/protobuf/encoding/protojson"
)

// Handler is the signature of the event callback
type Handler = func(event *asset.Event) error

// Listener listens to Chaincode events
type Listener struct {
	contract     *gateway.Contract
	registration fab.Registration
	ccEvents     <-chan *fab.CCEvent
	handler      Handler
	channel      string
	logger       zerolog.Logger

	startTxID    string
	startEventID string
}

type ListenerChaincodeData struct {
	Wallet    *Wallet
	Config    core.ConfigProvider
	MSPID     string
	Channel   string
	Chaincode string
}

// connectToGateway connects to the Fabric gateway using the provided ListenerChaincodeData
func connectToGateway(ccData *ListenerChaincodeData, options ...gateway.Option) (*gateway.Gateway, error) {
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

func extractTxIDFromEventID(id string) (string, error) {
	s := strings.Split(id, orcledger.EventIDSeparator)

	if len(s) != 2 {
		return "", errors.New("cannot extract TxID from event id string")
	}

	return s[0], nil
}

func queryBlockNumberByTxID(txID string, ccData *ListenerChaincodeData) (uint64, error) {
	gw, err := connectToGateway(ccData)
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

// NewListener instanciates a Listener listening events on the configured blockchain.
// It filters only events emitted by the chaincode (see ledger.EventName).
// The handler callback will be called for every event received.
func NewListener(ccData *ListenerChaincodeData, startEventID string, handler Handler) (*Listener, error) {
	logger := log.With().
		Str("channel", ccData.Channel).
		Str("chaincode", ccData.Chaincode).
		Logger()

	var err error

	startTxID := ""
	blockNum := uint64(0)
	if startEventID != "" {
		startTxID, err = extractTxIDFromEventID(startEventID)
		if err != nil {
			return nil, err
		}

		blockNum, err = queryBlockNumberByTxID(startTxID, ccData)
		if err != nil {
			return nil, err
		}
	}

	logger.Debug().Uint64("startEventBlock", blockNum).Str("startTxID", startTxID).Msg("instanciating event listener")

	gw, err := connectToGateway(ccData, gateway.WithBlockNum(blockNum))
	if err != nil {
		return nil, err
	}
	defer gw.Close()

	network, err := gw.GetNetwork(ccData.Channel)
	if err != nil {
		return nil, err
	}

	contract := network.GetContract(ccData.Chaincode)

	registration, ccEventStream, err := contract.RegisterEvent(orcledger.EventName)
	if err != nil {
		return nil, err
	}

	return &Listener{
		contract:     contract,
		registration: registration,
		ccEvents:     ccEventStream,
		handler:      handler,
		channel:      ccData.Channel,
		logger:       logger,
		startEventID: startEventID,
		startTxID:    startTxID,
	}, nil
}

// Close will unregister the chaincode listener
func (l *Listener) Close() {
	l.logger.Debug().Msg("Closing chaincode event listener")
	l.contract.Unregister(l.registration)
}

// Listen will trigger the callback with every event received
func (l *Listener) Listen(ctx context.Context) error {
	// As a block may have multiple transactions, make sure we skip ccEvents until we reach the last seen
	skipCCEvent := l.startTxID != ""
	// As a transaction may have multiple events, make sure we skip events until we reach the last seen
	skipEvent := l.startEventID != ""

	var err error
	for {
		select {
		case ccEvent := <-l.ccEvents:
			logger := l.logger.With().
				Str("eventName", ccEvent.EventName).
				Uint64("blockNumber", ccEvent.BlockNumber).
				Str("source", ccEvent.SourceURL).
				Str("txID", ccEvent.TxID).
				Logger()

			skipCCEvent = skipCCEvent && ccEvent.TxID != l.startTxID
			if skipCCEvent {
				logger.Debug().Msg("skipping previously handled ccEvent")
				break
			}

			logger.Debug().Msg("handling ccEvent")
			skipEvent, err = l.handleCCEvent(ccEvent, skipEvent)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			l.logger.Debug().Msg("context done: stop listening")
			return ctx.Err()
		}
	}
}

func (l *Listener) handleCCEvent(ccEvent *fab.CCEvent, skipEvent bool) (bool, error) {
	var rawEvents []json.RawMessage
	err := json.Unmarshal(ccEvent.Payload, &rawEvents)
	if err != nil {
		l.logger.Error().Err(err).Str("payload", string(ccEvent.Payload)).Msg("failed to deserialize chaincode transaction")
		return false, err
	}

	for _, rawEvent := range rawEvents {
		event := new(asset.Event)
		err = protojson.Unmarshal(rawEvent, event)
		if err != nil {
			l.logger.Error().Str("rawEvent", string(rawEvent)).Err(err).Msg("failed to deserialize event")
			return false, err
		}

		skipEvent = skipEvent && event.Id != l.startEventID
		if skipEvent || event.Id == l.startEventID {
			l.logger.Debug().Interface("event", event).Msg("skipping previously handled event")
			continue
		}

		event.Channel = l.channel

		l.logger.Debug().Interface("event", event).Msg("handling event")
		err = l.handler(event)
		if err != nil {
			return false, err
		}
	}

	return skipEvent, nil
}
