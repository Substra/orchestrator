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
	"google.golang.org/grpc"
)

type AppServer struct {
	grpc *grpc.Server
	amqp *common.Session
	db   *Database
}

func GetServer(dbURL string, rabbitDSN string, additionalOptions []grpc.ServerOption) (*AppServer, error) {
	pgDB, err := InitDatabase(dbURL)
	if err != nil {
		return nil, err
	}

	session := common.NewSession("orchestrator", rabbitDSN)

	// providerInterceptor will wrap gRPC requests and inject a ServiceProvider in request's context
	providerInterceptor := NewProviderInterceptor(pgDB, session)
	concurrencyLimiter := new(ConcurrencyLimiter)

	interceptor := grpc.ChainUnaryInterceptor(
		common.LogRequest,
		common.InterceptStandaloneErrors,
		concurrencyLimiter.Intercept,
		common.InterceptMSPID,
		common.InterceptChannel,
		providerInterceptor.Intercept,
	)

	serverOptions := append(additionalOptions, interceptor)

	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterNodeServiceServer(server, NewNodeServer())
	asset.RegisterObjectiveServiceServer(server, NewObjectiveServer())
	asset.RegisterDataSampleServiceServer(server, NewDataSampleServer())
	asset.RegisterAlgoServiceServer(server, NewAlgoServer())
	asset.RegisterDataManagerServiceServer(server, NewDataManagerServer())
	asset.RegisterDatasetServiceServer(server, NewDatasetServer())
	asset.RegisterComputeTaskServiceServer(server, NewComputeTaskServer())
	asset.RegisterModelServiceServer(server, NewModelServer())
	asset.RegisterComputePlanServiceServer(server, NewComputePlanServer())

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
