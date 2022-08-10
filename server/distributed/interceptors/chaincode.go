package interceptors

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/substra/orchestrator/chaincode/contracts"
	"github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/common/logger"
	"github.com/substra/orchestrator/server/distributed/chaincode"
	"google.golang.org/grpc"
)

// InvocatorInterceptor is a structure referencing a wallet which will keeps identities for the duration of program execution.
// It is responsible for injecting a chaincode Invocator in the request's context.
type InvocatorInterceptor struct {
	gwPool chaincode.GatewayPool
}

// NewInvocatorInterceptor creates an InvocatorInterceptor
func NewInvocatorInterceptor(config core.ConfigProvider, wallet *chaincode.Wallet, gatewayTimeout time.Duration) *InvocatorInterceptor {
	checker := contracts.NewContractCollection()
	return &InvocatorInterceptor{
		gwPool: chaincode.NewGatewayPool(config, wallet, gatewayTimeout, checker),
	}
}

// Close make sure all gateways are closed
func (ci *InvocatorInterceptor) Close() {
	ci.gwPool.Close()
}

// UnaryServerInterceptor is a gRPC interceptor and will make the fabric contract available in the request context
func (ci *InvocatorInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range interceptors.IgnoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	mspid, err := interceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}

	channel, err := interceptors.ExtractChannel(ctx)
	if err != nil {
		return nil, err
	}

	chaincodeName, err := extractChaincodeName(ctx)
	if err != nil {
		return nil, err
	}

	logger.Get(ctx).WithField("chaincode", chaincodeName).
		WithField("channel", channel).
		WithField("mspid", mspid).
		Debug("Successfully retrieved chaincode metadata from headers")

	gw, err := ci.gwPool.GetGateway(ctx, mspid)
	if err != nil {
		return nil, err
	}

	invocator := chaincode.NewContractInvocator(gw, channel, chaincodeName)

	newCtx := WithInvocator(ctx, invocator)
	return handler(newCtx, req)
}

type ctxInvocatorMarker struct{}

var ctxInvocatorKey = &ctxInvocatorMarker{}

func WithInvocator(ctx context.Context, invocator chaincode.Invocator) context.Context {
	return context.WithValue(ctx, ctxInvocatorKey, invocator)
}

// ExtractInvocator retrieves chaincode Invocator from gRPC context
// Invocator is expected to be set by InvocatorInterceptor.UnaryServerInterceptor()
func ExtractInvocator(ctx context.Context) (chaincode.Invocator, error) {
	invocator, ok := ctx.Value(ctxInvocatorKey).(chaincode.Invocator)
	if !ok {
		return nil, errors.New("invocator not found in context")
	}
	return invocator, nil
}
