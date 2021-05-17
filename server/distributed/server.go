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
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"google.golang.org/grpc"
)

type AppServer struct {
	grpc *grpc.Server
}

func GetServer(networkConfig string, certificate string, key string, additionalOptions []grpc.ServerOption, orcConf *common.OrchestratorConfiguration) (*AppServer, error) {
	wallet := wallet.New(certificate, key)
	config := config.FromFile(networkConfig)

	chaincodeInterceptor, err := NewInterceptor(config, wallet)
	if err != nil {
		return nil, err

	}

	channelInterceptor := common.NewChannelInterceptor(orcConf)

	interceptor := grpc.ChainUnaryInterceptor(
		common.LogRequest,
		common.InterceptDistributedErrors,
		common.InterceptMSPID,
		channelInterceptor.InterceptChannel,
		chaincodeInterceptor.Intercept,
	)

	serverOptions := append(additionalOptions, interceptor)

	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterNodeServiceServer(server, NewNodeAdapter())
	asset.RegisterObjectiveServiceServer(server, NewObjectiveAdapter())
	asset.RegisterDataSampleServiceServer(server, NewDataSampleAdapter())
	asset.RegisterAlgoServiceServer(server, NewAlgoAdapter())
	asset.RegisterDataManagerServiceServer(server, NewDataManagerAdapter())
	asset.RegisterDatasetServiceServer(server, NewDatasetAdapter())
	asset.RegisterComputeTaskServiceServer(server, NewComputeTaskAdapter())
	asset.RegisterModelServiceServer(server, NewModelAdapter())
	asset.RegisterComputePlanServiceServer(server, NewComputePlanAdapter())
	asset.RegisterPerformanceServiceServer(server, NewPerformanceAdapter())

	return &AppServer{server}, nil
}

func (a *AppServer) GetGrpcServer() *grpc.Server {
	return a.grpc
}

func (a *AppServer) Stop() {
	a.grpc.Stop()
}
