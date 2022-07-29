package interceptors

import (
	"context"
	"errors"
	"fmt"
	"github.com/owkin/orchestrator/server/common"
	"google.golang.org/grpc/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/owkin/orchestrator/forwarder/event"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"google.golang.org/grpc"
)

type ChaincodeDataInterceptor struct {
	wallet *wallet.Wallet
	config core.ConfigProvider
}

func NewChaincodeDataInterceptor(
	wallet *wallet.Wallet,
	config core.ConfigProvider,
) *ChaincodeDataInterceptor {
	return &ChaincodeDataInterceptor{
		wallet: wallet,
		config: config,
	}
}

const headerChaincode = "chaincode"

func (i *ChaincodeDataInterceptor) StreamServerInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := stream.Context()

	channel, err := common.ExtractChannel(ctx)
	if err != nil {
		return nil
	}

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return err
	}

	// TODO: factor metadata retrieval with
	// 	server/distributed/interceptors.go:64
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("could not extract metadata")
	}

	if len(md.Get(headerChaincode)) != 1 {
		return fmt.Errorf("missing or invalid header '%s'", headerChaincode)
	}

	chaincode := md.Get(headerChaincode)[0]

	ccData := &event.ListenerChaincodeData{
		Wallet:    i.wallet,
		Config:    i.config,
		MSPID:     mspid,
		Channel:   channel,
		Chaincode: chaincode,
	}
	newCtx := WithChaincodeData(ctx, ccData)
	streamWithContext := common.BindStreamToContext(newCtx, stream)

	return handler(srv, streamWithContext)
}

type ctxChaincodeDataInterceptorMarker struct{}

var ctxChaincodeDataKey = &ctxChaincodeDataInterceptorMarker{}

func WithChaincodeData(ctx context.Context, ccData *event.ListenerChaincodeData) context.Context {
	return context.WithValue(ctx, ctxChaincodeDataKey, ccData)
}

func ExtractChaincodeData(ctx context.Context) (*event.ListenerChaincodeData, error) {
	ccData, ok := ctx.Value(ctxChaincodeDataKey).(*event.ListenerChaincodeData)
	if !ok {
		return nil, errors.New("chaincode data not found in context")
	}
	return ccData, nil
}
