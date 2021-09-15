// Package gateway contains the abstraction over fabric's gateway API.
// It exposes structures to reuse the gateway in long-running processes.
package gateway

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/distributed/wallet"
)

// ChaincodeRequester describe a component capable of querying the chaincode.
type ChaincodeRequester interface {
	Request(ctx context.Context, channel, chaincode, method string, args []byte) (<-chan []byte, <-chan error)
}

// chaincodeRequest is a request to the chaincode
type chaincodeRequest struct {
	ctx       context.Context
	channel   string
	chaincode string
	method    string
	args      []byte
	out       chan []byte
	err       chan error
}

// Gateway wraps fabric Gateway
type Gateway struct {
	gw       *gateway.Gateway
	requests chan chaincodeRequest
	closed   bool
	config   core.ConfigProvider
	checker  common.TransactionChecker
	mspid    string
}

// Pool is a container for multiple gateways.
type Pool struct {
	wallet   *wallet.Wallet
	config   core.ConfigProvider
	timeout  time.Duration
	gwLock   *sync.RWMutex
	gateways map[string]Gateway
	checker  common.TransactionChecker
}

// NewPool creates an empty Pool.
func NewPool(config core.ConfigProvider, wallet *wallet.Wallet, gatewayTimeout time.Duration, checker common.TransactionChecker) Pool {
	return Pool{
		wallet:   wallet,
		config:   config,
		timeout:  gatewayTimeout,
		gwLock:   new(sync.RWMutex),
		gateways: make(map[string]Gateway),
		checker:  checker,
	}
}

// Close will close all gateways in the Pool.
func (p *Pool) Close() {
	p.gwLock.RLock()
	defer p.gwLock.RUnlock()

	for _, gw := range p.gateways {
		gw.Close()
	}
}

// GetGateway returns the existing gateway for the given MSP ID or creates one if needed.
func (p *Pool) GetGateway(ctx context.Context, mspid string) (*Gateway, error) {
	label := mspid + "-id"
	log := logger.Get(ctx).WithField("mspid", mspid)

	p.gwLock.RLock()
	gw, ok := p.gateways[label]
	p.gwLock.RUnlock()
	if ok {
		return &gw, nil
	}

	// Gateway does not exist yet, let's create one

	p.gwLock.Lock()
	defer p.gwLock.Unlock()

	log.Debug("creating new gateway connection")

	err := p.wallet.EnsureIdentity(label, mspid)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	fabricGw, err := gateway.Connect(
		gateway.WithConfig(p.config),
		gateway.WithIdentity(p.wallet, label),
		gateway.WithTimeout(p.timeout),
	)
	if err != nil {
		return nil, err
	}

	gw = p.newGateway(fabricGw, mspid)
	p.gateways[label] = gw

	elapsed := time.Since(start)
	log.WithField("duration", elapsed).Debug("Connected to gateway")

	return &gw, nil
}

// newGateway instanciates a Gateway object and spawn a goroutine to process chaincode requests.
func (p *Pool) newGateway(fabricGateway *gateway.Gateway, mspid string) Gateway {
	gateway := Gateway{
		gw:       fabricGateway,
		requests: make(chan chaincodeRequest),
		mspid:    mspid,
		checker:  p.checker,
		config:   p.config,
		closed:   false,
	}

	go gateway.processRequests()

	return gateway
}

// Close will properly shutdown chaincode requests processing.
func (gw *Gateway) Close() {
	gw.closed = true
	close(gw.requests)
	// This is a noop right now but fabric code mentions a future cleanup
	gw.gw.Close()
}

// processRequests loops over chaincode requests
func (gw *Gateway) processRequests() {
	for request := range gw.requests {
		gw.invoke(request)
	}
	log.WithField("mspip", gw.mspid).Info("Stopping chaincode request processing")
}

// invoke the chaincode according to the request and send results over the channels
func (gw *Gateway) invoke(req chaincodeRequest) {
	defer close(req.err)
	defer close(req.out)

	isEvaluate := gw.checker.IsEvaluateMethod(req.method)

	log := logger.Get(req.ctx).
		WithFields(
			log.F("mspid", gw.mspid),
			log.F("channel", req.channel),
			log.F("chaincode", req.chaincode),
			log.F("method", req.method),
			log.F("evaluate", isEvaluate),
		)

	log.Debug("Calling chaincode")

	configBackend, err := gw.config()
	if err != nil {
		req.err <- err
		return
	}
	peers, err := extractChannelLocalPeers(configBackend, req.channel)
	if err != nil {
		req.err <- err
		return
	}

	network, err := gw.gw.GetNetwork(req.channel)
	if err != nil {
		req.err <- err
		return
	}

	contract := network.GetContract(req.chaincode)

	var data []byte

	if isEvaluate {
		var tx *gateway.Transaction
		tx, err = contract.CreateTransaction(req.method, gateway.WithEndorsingPeers(peers...))

		if err != nil {
			req.err <- err
			return
		}

		data, err = tx.Evaluate(string(req.args))
	} else {
		data, err = contract.SubmitTransaction(req.method, string(req.args))
	}

	if err != nil {
		req.err <- err
		return
	}

	req.out <- data
}

// Request will create a chaincode request with given inputs.
// This request will be processed asynchronously by a gateway and results will be available through returned channels.
func (gw *Gateway) Request(ctx context.Context, channel, chaincode, method string, args []byte) (<-chan []byte, <-chan error) {
	// buffered channels since we may send a response while the context has already been canceled.
	// This avoids blocking if nothing reads the channels.
	out := make(chan []byte, 1)
	err := make(chan error, 1)

	if gw.closed {
		logger.Get(ctx).Warn("Gateway closed")
		err <- orcerrors.NewInternal("gateway closed")
		close(out)
		close(err)
		return out, err
	}

	request := chaincodeRequest{
		ctx:       ctx,
		channel:   channel,
		chaincode: chaincode,
		method:    method,
		args:      args,
		out:       out,
		err:       err,
	}

	gw.requests <- request

	return out, err
}

// ExtractChannelLocalPeers retrieves the local peers present in the provided channel from the config file
func extractChannelLocalPeers(configBackend []core.ConfigBackend, channel string) ([]string, error) {
	if len(configBackend) != 1 {
		return nil, errors.New("invalid config file")
	}

	config := configBackend[0]
	channelPeers, _ := config.Lookup(fmt.Sprintf("channels.%s.peers", channel))

	peersMap, ok := channelPeers.(map[string]interface{})

	if !ok {
		return nil, errors.New("invalid config structure")
	}

	peers := make([]string, 0, len(peersMap))
	for k := range peersMap {
		peers = append(peers, k)
	}
	return peers, nil
}
