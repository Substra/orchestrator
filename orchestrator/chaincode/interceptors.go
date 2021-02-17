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

package chaincode

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/orchestrator/chaincode/wallet"
	"github.com/owkin/orchestrator/orchestrator/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const headerChaincode = "chaincode"
const headerChannel = "channel"

type chaincodeMetadata struct {
	channel   string
	chaincode string
}

var requiredHeaders = [...]string{headerChaincode, headerChannel}
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

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("Could not extract metadata")
	}

	ccMetadata, err := getCCMetadata(md)
	if err != nil {
		return nil, err
	}
	log.WithField("ccMetadata", ccMetadata).Debug("Successfully retrieved chaincode metadata from headers")

	label := mspid + "-id"

	ci.wallet.EnsureIdentity(label, mspid)
	start := time.Now()
	gw, err := gateway.Connect(
		gateway.WithConfig(ci.config),
		gateway.WithIdentity(ci.wallet, label),
	)
	elapsed := time.Since(start)
	log.WithField("duration", elapsed).Debug("Connected to gateway")

	if err != nil {
		return nil, err
	}

	defer gw.Close()

	network, err := gw.GetNetwork(ccMetadata.channel)
	if err != nil {
		return nil, err
	}

	contract := network.GetContract(ccMetadata.chaincode)
	invocator := NewContractInvocator(contract)

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
		return nil, errors.New("Invocator not found in context")
	}
	return invocator, nil
}

// getCCMetadata make sure all necessary headers are set and returns
// chaincode-related metadata
func getCCMetadata(md metadata.MD) (*chaincodeMetadata, error) {
	for _, h := range requiredHeaders {
		values := md.Get(h)
		if len(values) != 1 {
			return nil, fmt.Errorf("Missing or invalid header '%s'", h)
		}
	}

	return &chaincodeMetadata{
		channel:   md.Get(headerChannel)[0],
		chaincode: md.Get(headerChaincode)[0],
	}, nil
}
