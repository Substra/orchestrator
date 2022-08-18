package interceptors

import (
	"context"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/substra/orchestrator/server/common"
	"github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/distributed/chaincode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const headerChaincode = "chaincode"

func extractChaincodeName(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("could not extract metadata")
	}

	if len(md.Get(headerChaincode)) != 1 {
		return "", fmt.Errorf("missing or invalid header '%s'", headerChaincode)
	}

	return md.Get(headerChaincode)[0], nil
}

type ChaincodeDataInterceptor struct {
	wallet *chaincode.Wallet
	config core.ConfigProvider
}

func NewChaincodeDataInterceptor(
	wallet *chaincode.Wallet,
	config core.ConfigProvider,
) *ChaincodeDataInterceptor {
	return &ChaincodeDataInterceptor{
		wallet: wallet,
		config: config,
	}
}

func (i *ChaincodeDataInterceptor) StreamServerInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := stream.Context()

	channel, err := interceptors.ExtractChannel(ctx)
	if err != nil {
		return nil
	}

	mspid, err := interceptors.ExtractMSPID(ctx)
	if err != nil {
		return err
	}

	chaincodeName, err := extractChaincodeName(ctx)
	if err != nil {
		return err
	}

	ccData := &chaincode.ListenerChaincodeData{
		Wallet:    i.wallet,
		Config:    i.config,
		MSPID:     mspid,
		Channel:   channel,
		Chaincode: chaincodeName,
	}
	newCtx := WithChaincodeData(ctx, ccData)
	streamWithContext := common.BindStreamToContext(newCtx, stream)

	return handler(srv, streamWithContext)
}

type ctxChaincodeDataInterceptorMarker struct{}

var ctxChaincodeDataKey = &ctxChaincodeDataInterceptorMarker{}

func WithChaincodeData(ctx context.Context, ccData *chaincode.ListenerChaincodeData) context.Context {
	return context.WithValue(ctx, ctxChaincodeDataKey, ccData)
}

func ExtractChaincodeData(ctx context.Context) (*chaincode.ListenerChaincodeData, error) {
	ccData, ok := ctx.Value(ctxChaincodeDataKey).(*chaincode.ListenerChaincodeData)
	if !ok {
		return nil, errors.New("chaincode data not found in context")
	}
	return ccData, nil
}
