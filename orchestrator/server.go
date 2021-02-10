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
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/orchestration"
	"github.com/owkin/orchestrator/orchestrator/chaincode"
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
		log.Fatalf("Missing environment variable: %s", n)
	}
	return v
}

// RunServerWithChainCode is exported
func RunServerWithChainCode() {
	chaincodeInterceptor, err := chaincode.NewInterceptor(mustGetEnv("NETWORK_CONFIG"), mustGetEnv("CERT"), mustGetEnv("KEY"))
	if err != nil {
		log.Fatalf("Failed to instanciate chaincode interceptor: %v", err)

	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(common.InterceptMSPID, chaincodeInterceptor.Intercept))

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
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	log.WithField("address", listen.Addr().String()).Info("Server listening")
	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to server grpc server on port 9000: %v", err)
	}
}

// RunServerWithoutChainCode will expose the chaincode logic through gRPC.
// State will be stored in a redis database.
func RunServerWithoutChainCode() {
	dsn := os.Getenv("COUCHDB_DSN")
	if dsn == "" {
		dsn = "http://dev:dev@localhost:5984"
	}
	couchPersistence, err := standalone.NewPersistence(context.TODO(), dsn, "substra_orchestrator")
	defer couchPersistence.Close(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	provider := orchestration.NewServiceProvider(couchPersistence)

	server := grpc.NewServer(grpc.UnaryInterceptor(common.InterceptMSPID))

	// Register application services
	assets.RegisterNodeServiceServer(server, standalone.NewNodeServer(provider.GetNodeService()))
	assets.RegisterObjectiveServiceServer(server, standalone.NewObjectiveServer(provider.GetObjectiveService()))

	// Register reflection service
	reflection.Register(server)

	// Register healthcheck service
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(server, healthcheck)

	log.WithField("address", listen.Addr().String()).Info("Server listening")
	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to server grpc server on port 9000: %v", err)
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
