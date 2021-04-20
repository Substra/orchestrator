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

package distributed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/chaincode/contracts"
	"github.com/owkin/orchestrator/server/common"
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
		return nil, errors.New("Could not extract metadata")
	}

	if len(md.Get(headerChaincode)) != 1 {
		return nil, fmt.Errorf("Missing or invalid header '%s'", headerChaincode)
	}

	chaincode := md.Get(headerChaincode)[0]

	log.WithField("chaincode", chaincode).
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
	log.WithField("duration", elapsed).Debug("Connected to gateway")

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
	invocator := NewContractInvocator(contract, checker)

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
