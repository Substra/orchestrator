package distributed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/chaincode/contracts"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const headerChaincode = "chaincode"

var ignoredMethods = [...]string{
	"grpc.health",
}

// Interceptor is a structure referencing a wallet which will keeps identities for the duration of program execution.
// It is responsible for injecting a chaincode Invocator in the request's context.
type Interceptor struct {
	wallet *wallet.Wallet
	config core.ConfigProvider
}

// NewInterceptor creates an Interceptor
func NewInterceptor(config core.ConfigProvider, wallet *wallet.Wallet) (*Interceptor, error) {
	return &Interceptor{wallet: wallet, config: config}, nil
}

// Intercept is a gRPC interceptor and will make the fabric contract available in the request context
func (ci *Interceptor) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}

	channel, err := common.ExtractChannel(ctx)
	if err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("could not extract metadata")
	}

	if len(md.Get(headerChaincode)) != 1 {
		return nil, fmt.Errorf("missing or invalid header '%s'", headerChaincode)
	}

	chaincode := md.Get(headerChaincode)[0]

	logger.Get(ctx).WithField("chaincode", chaincode).
		WithField("channel", channel).
		WithField("mspid", mspid).
		Debug("Successfully retrieved chaincode metadata from headers")

	label := mspid + "-id"

	err = ci.wallet.EnsureIdentity(label, mspid)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	gw, err := gateway.Connect(
		gateway.WithConfig(ci.config),
		gateway.WithIdentity(ci.wallet, label),
	)
	elapsed := time.Since(start)
	logger.Get(ctx).WithField("duration", elapsed).Debug("Connected to gateway")

	if err != nil {
		return nil, err
	}

	configBackend, err := ci.config()
	if err != nil {
		return nil, err
	}

	peers, err := extractChannelLocalPeers(configBackend, channel)
	if err != nil {
		return nil, err
	}

	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		return nil, err
	}

	contract := network.GetContract(chaincode)
	checker := contracts.NewContractCollection()
	invocator := NewContractInvocator(contract, checker, peers)

	newCtx := context.WithValue(ctx, ctxInvocatorKey, invocator)
	return handler(newCtx, req)
}

type ctxInvocatorMarker struct{}

var (
	ctxInvocatorKey = &ctxInvocatorMarker{}
)

// ExtractInvocator retrieves chaincode Invocator from gRPC context
// Invocator is expected to be set by ChaincodeInterceptor.Intercept()
func ExtractInvocator(ctx context.Context) (Invocator, error) {
	invocator, ok := ctx.Value(ctxInvocatorKey).(Invocator)
	if !ok {
		return nil, errors.New("invocator not found in context")
	}
	return invocator, nil
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
