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

package standalone

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/standalone/dbal"
	"github.com/owkin/orchestrator/server/standalone/handlers"
	"github.com/owkin/orchestrator/server/standalone/interceptors"
	"google.golang.org/grpc"
)

type AppServer struct {
	grpc *grpc.Server
	amqp *common.Session
	db   *dbal.Database
}

func GetServer(dbURL string, rabbitDSN string, additionalOptions []grpc.ServerOption, config *common.OrchestratorConfiguration) (*AppServer, error) {
	pgDB, err := dbal.InitDatabase(dbURL)
	if err != nil {
		return nil, err
	}

	session := common.NewSession("orchestrator", rabbitDSN)

	channelInterceptor := common.NewChannelInterceptor(config)
	// providerInterceptor will wrap gRPC requests and inject a ServiceProvider in request's context
	providerInterceptor := interceptors.NewProviderInterceptor(pgDB, session)

	interceptor := grpc.ChainUnaryInterceptor(
		common.LogRequest,
		common.InterceptStandaloneErrors,
		common.InterceptMSPID,
		channelInterceptor.InterceptChannel,
		providerInterceptor.Intercept,
	)

	serverOptions := append(additionalOptions, interceptor)

	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterNodeServiceServer(server, handlers.NewNodeServer())
	asset.RegisterObjectiveServiceServer(server, handlers.NewObjectiveServer())
	asset.RegisterDataSampleServiceServer(server, handlers.NewDataSampleServer())
	asset.RegisterAlgoServiceServer(server, handlers.NewAlgoServer())
	asset.RegisterDataManagerServiceServer(server, handlers.NewDataManagerServer())
	asset.RegisterDatasetServiceServer(server, handlers.NewDatasetServer())
	asset.RegisterComputeTaskServiceServer(server, handlers.NewComputeTaskServer())
	asset.RegisterModelServiceServer(server, handlers.NewModelServer())
	asset.RegisterComputePlanServiceServer(server, handlers.NewComputePlanServer())
	asset.RegisterPerformanceServiceServer(server, handlers.NewPerformanceServer())
	asset.RegisterEventServiceServer(server, handlers.NewEventServer())

	return &AppServer{
		grpc: server,
		amqp: session,
		db:   pgDB,
	}, nil
}

func (a *AppServer) GetGrpcServer() *grpc.Server {
	return a.grpc
}

func (a *AppServer) Stop() {
	a.grpc.Stop()
	a.db.Close()
	a.amqp.Close()
}
