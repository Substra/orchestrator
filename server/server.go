// Copyright 2020 Owkin Inc.
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

// server binary exposing a gRPC interface to manage distributed learning asset.
// It can run in either standalone or distributed mode.
// In standalone mode it handle all the logic while in distributed mode everything is delegated to a chaincode.
package main

import (
	"flag"
	"net"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/distributed"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"github.com/owkin/orchestrator/server/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func runDistributed() {
	networkConfig := common.MustGetEnv("NETWORK_CONFIG")

	wallet := wallet.New(common.MustGetEnv("FABRIC_CERT"), common.MustGetEnv("FABRIC_KEY"))

	config := config.FromFile(networkConfig)

	var serverOptions []grpc.ServerOption

	chaincodeInterceptor, err := distributed.NewInterceptor(config, wallet)
	if err != nil {
		log.WithError(err).Fatal("Failed to instanciate chaincode interceptor")
	}

	interceptor := grpc.ChainUnaryInterceptor(
		common.LogRequest,
		common.InterceptMSPID,
		common.InterceptChannel,
		chaincodeInterceptor.Intercept,
	)

	serverOptions = append(serverOptions, interceptor)

	if tlsOptions := getTLSOptions(); tlsOptions != nil {
		serverOptions = append(serverOptions, tlsOptions)
	}

	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterNodeServiceServer(server, distributed.NewNodeAdapter())
	asset.RegisterObjectiveServiceServer(server, distributed.NewObjectiveAdapter())

	// Register reflection service
	reflection.Register(server)

	// Register healthcheck service
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(server, healthcheck)

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.WithError(err).Fatal("failed to listen on port 9000")
	}

	log.WithField("address", listen.Addr().String()).Info("Server listening")
	if err := server.Serve(listen); err != nil {
		log.WithError(err).Fatal("failed to server grpc server on port 9000")
	}
}

// runStandalone will expose the chaincode logic through gRPC.
func runStandalone() {
	dbURL := common.MustGetEnv("DATABASE_URL")
	pgDB, err := standalone.InitDatabase(dbURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to create persistence layer")
	}
	defer pgDB.Close()

	rabbitDSN := common.MustGetEnv("AMQP_DSN")
	session := common.NewSession("orchestrator", rabbitDSN)
	defer session.Close()

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.WithError(err).Fatal("failed to listen on port 9000")
	}

	var serverOptions []grpc.ServerOption

	// providerInterceptor will wrap gRPC requests and inject a ServiceProvider in request's context
	providerInterceptor := standalone.NewProviderInterceptor(pgDB, session)
	concurrencyLimiter := new(standalone.ConcurrencyLimiter)
	interceptor := grpc.ChainUnaryInterceptor(
		common.LogRequest,
		common.InterceptErrors,
		concurrencyLimiter.Intercept,
		common.InterceptMSPID,
		common.InterceptChannel,
		providerInterceptor.Intercept,
	)
	serverOptions = append(serverOptions, interceptor)

	// TLS
	if tlsOptions := getTLSOptions(); tlsOptions != nil {
		serverOptions = append(serverOptions, tlsOptions)
	}

	// Declare GRPC server
	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterNodeServiceServer(server, standalone.NewNodeServer())
	asset.RegisterObjectiveServiceServer(server, standalone.NewObjectiveServer())

	// Register reflection service
	reflection.Register(server)

	// Register healthcheck service
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(server, healthcheck)

	log.WithField("address", listen.Addr().String()).Info("Server listening")
	if err := server.Serve(listen); err != nil {
		log.WithError(err).Fatal("failed to server grpc server on port 9000")
	}
}

func main() {
	var standaloneMode = false

	common.InitLogging()

	flag.BoolVar(&standaloneMode, "standalone", true, "Run the chaincode in standalone mode")
	flag.BoolVar(&standaloneMode, "s", true, "Run the chaincode in standalone mode (shorthand)")

	flag.Parse()

	mode, ok := common.GetEnv("MODE")
	if ok {
		switch mode {
		case "chaincode":
			standaloneMode = false
		case "standalone":
			standaloneMode = true
		}
	}

	if standaloneMode {
		runStandalone()
	} else {
		runDistributed()
	}
}
