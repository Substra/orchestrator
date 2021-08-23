package distributed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/owkin/orchestrator/chaincode/contracts"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/distributed/gateway"
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
	gwPool gateway.Pool
}

// NewInterceptor creates an Interceptor
func NewInterceptor(config core.ConfigProvider, wallet *wallet.Wallet, gatewayTimeout time.Duration) *Interceptor {
	checker := contracts.NewContractCollection()
	return &Interceptor{
		gwPool: gateway.NewPool(config, wallet, gatewayTimeout, checker),
	}
}

// Close make sure all gateways are closed
func (ci *Interceptor) Close() {
	ci.gwPool.Close()
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

	gw, err := ci.gwPool.GetGateway(ctx, mspid)
	if err != nil {
		return nil, err
	}

	invocator := NewContractInvocator(gw, channel, chaincode)

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
