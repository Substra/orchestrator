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

package main

import (
	"context"
	"flag"
	"net"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/orchestrator/chaincode"
	chaincodeEvents "github.com/owkin/orchestrator/orchestrator/chaincode/events"
	"github.com/owkin/orchestrator/orchestrator/chaincode/wallet"
	"github.com/owkin/orchestrator/orchestrator/common"
	"github.com/owkin/orchestrator/orchestrator/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// envPrefix is the string prefixing environment variables related to the orchestrator
const envPrefix = "ORCHESTRATOR_"

// Whether to run in standalone mode or not
var standaloneMode = false

// mustGetEnv extract environment variable or abort with an error message
// Every env var is prefixed by ORCHESTRATOR_
func mustGetEnv(name string) string {
	n := envPrefix + name
	v, ok := os.LookupEnv(n)
	if !ok {
		log.WithField("env_var", n).Fatal("Missing environment variable")
	}
	return v
}

// RunServerWithChainCode is exported
func RunServerWithChainCode() {
	networkConfig := mustGetEnv("NETWORK_CONFIG")

	rabbitDSN := mustGetEnv("AMQP_DSN")
	session := common.NewSession("orchestrator", rabbitDSN)
	defer session.Close()

	wallet := wallet.New(mustGetEnv("CERT"), mustGetEnv("KEY"))

	config := config.FromFile(networkConfig)

	converter := chaincodeEvents.NewForwarder(session)

	listener, err := chaincodeEvents.NewListener(
		wallet,
		config,
		mustGetEnv("MSPID"),
		mustGetEnv("CHANNEL"),
		mustGetEnv("CHAINCODE"),
		converter.Forward,
	)

	if err != nil {
		log.WithError(err).Fatal("Failed to instanciate listener")
	}

	defer listener.Close()
	go listener.Listen()

	chaincodeInterceptor, err := chaincode.NewInterceptor(config, wallet)
	if err != nil {
		log.WithError(err).Fatal("Failed to instanciate chaincode interceptor")

	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		common.LogRequest,
		common.InterceptMSPID,
		chaincodeInterceptor.Intercept,
	))

	// Register application services
	assets.RegisterNodeServiceServer(server, chaincode.NewNodeAdapter())
	assets.RegisterObjectiveServiceServer(server, chaincode.NewObjectiveAdapter())

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

// RunServerWithoutChainCode will expose the chaincode logic through gRPC.
// State will be stored in a redis database.
func RunServerWithoutChainCode() {
	couchDSN := os.Getenv("COUCHDB_DSN")
	if couchDSN == "" {
		couchDSN = "http://dev:dev@localhost:5984"
	}
	couchPersistence, err := standalone.NewPersistence(context.TODO(), couchDSN, "substra_orchestrator")
	defer couchPersistence.Close(context.TODO())
	if err != nil {
		log.WithError(err).Fatal("Failed to create persistence layer")
	}

	rabbitDSN := mustGetEnv("AMQP_DSN")
	session := common.NewSession("orchestrator", rabbitDSN)
	defer session.Close()

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.WithError(err).Fatal("failed to listen on port 9000")
	}

	// providerInterceptor will wrap gRPC requests and inject a ServiceProvider in request's context
	providerInterceptor := standalone.NewProviderInterceptor(couchPersistence, session)
	concurrencyLimiter := new(standalone.ConcurrencyLimiter)

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		common.LogRequest,
		concurrencyLimiter.Intercept,
		common.InterceptMSPID,
		providerInterceptor.Intercept,
	))

	// Register application services
	assets.RegisterNodeServiceServer(server, standalone.NewNodeServer())
	assets.RegisterObjectiveServiceServer(server, standalone.NewObjectiveServer())

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
	cLog := console.New(true)
	log.AddHandler(cLog, log.AllLevels...)

	flag.BoolVar(&standaloneMode, "standalone", true, "Run the chaincode in standalone mode")
	flag.BoolVar(&standaloneMode, "s", true, "Run the chaincode in standalone mode (shorthand)")

	flag.Parse()

	switch os.Getenv(envPrefix + "MODE") {
	case "chaincode":
		standaloneMode = false
	case "standalone":
		standaloneMode = true
	}

	if standaloneMode {
		RunServerWithoutChainCode()
	} else {
		RunServerWithChainCode()
	}
}
