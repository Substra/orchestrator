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

// Whether to run in standalone mode or not
var standalone = false

// RunServerWithChainCode is exported
func RunServerWithChainCode() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatal("failed to open wallet")
	}

	if !wallet.Exists("appClient") {
		cert, err := ioutil.ReadFile("./secrets/cert.pem")
		if err != nil {
			log.Fatal("failed to read peer cert")
		}

		key, err := ioutil.ReadFile("./secrets/key.pem")
		if err != nil {
			log.Fatal("failed to read key")
		}

		identity := gateway.NewX509Identity("MyOrg1MSP", string(cert), string(key))

		wallet.Put("appClient", identity)
	}

	// get config path

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile("./config.yml")),
		gateway.WithIdentity(wallet, "appClient"),
	)

	if err != nil {
		log.Fatalf("failed to instanciate gateway: %v", err)
	}

	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("failed to get network: %v", err)
	}

	contract := network.GetContract("mycc")
	result, err := contract.SubmitTransaction("RegisterNode", "1")
	if err != nil {
		log.Fatal("failed to invoke registration")
	}

	log.Debug(result)
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

	switch os.Getenv("ORCHESTRATOR_MODE") {
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
