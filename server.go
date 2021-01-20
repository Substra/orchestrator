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
	"io/ioutil"
	"net"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/database/couchdb"
	orchestratorGrpc "github.com/owkin/orchestrator/grpc"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/orchestration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const envPrefix = "ORCHESTRATOR_"
const defaultIdentity = "appClient"

// Whether to run in standalone mode or not
var standalone = false

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
	wallet := gateway.NewInMemoryWallet()

	if !wallet.Exists(defaultIdentity) {
		cert, err := ioutil.ReadFile(mustGetEnv("CERT"))
		if err != nil {
			log.Fatal("failed to read peer cert")
		}

		key, err := ioutil.ReadFile(mustGetEnv("KEY"))
		if err != nil {
			log.Fatal("failed to read key")
		}

		identity := gateway.NewX509Identity(mustGetEnv("MSPID"), string(cert), string(key))

		wallet.Put(defaultIdentity, identity)
	}

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(mustGetEnv("NETWORK_CONFIG"))),
		gateway.WithIdentity(wallet, defaultIdentity),
	)

	if err != nil {
		log.Fatalf("failed to instanciate gateway: %v", err)
	}

	defer gw.Close()

	network, err := gw.GetNetwork(mustGetEnv("CHANNEL"))
	if err != nil {
		log.Fatalf("failed to get network: %v", err)
	}

	contract := network.GetContract(mustGetEnv("CHAINCODE"))
	result, err := contract.SubmitTransaction("registerNode", "1")
	if err != nil {
		log.Fatalf("failed to invoke registration: %v", err)
	}

	log.Debug(string(result))
}

// RunServerWithoutChainCode will expose the chaincode logic through gRPC.
// State will be stored in a redis database.
func RunServerWithoutChainCode() {
	dsn := os.Getenv("COUCHDB_DSN")
	if dsn == "" {
		dsn = "http://dev:dev@localhost:5984"
	}
	couchPersistence, err := couchdb.NewPersistence(context.TODO(), dsn, "substra_orchestrator")
	defer couchPersistence.Close(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	provider := orchestration.NewServiceProvider(couchPersistence)

	server := grpc.NewServer()

	// Register application services
	assets.RegisterNodeServiceServer(server, orchestratorGrpc.NewNodeServer(provider.GetNodeService()))
	assets.RegisterObjectiveServiceServer(server, orchestratorGrpc.NewObjectiveServer(provider.GetObjectiveService()))

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

	flag.BoolVar(&standalone, "standalone", true, "Run the chaincode in standalone mode")
	flag.BoolVar(&standalone, "s", true, "Run the chaincode in standalone mode (shorthand)")

	flag.Parse()

	switch os.Getenv(envPrefix + "MODE") {
	case "chaincode":
		standalone = false
	case "standalone":
		standalone = true
	}

	if standalone {
		RunServerWithoutChainCode()
	} else {
		RunServerWithChainCode()
	}
}
