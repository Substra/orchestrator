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
	"io/ioutil"
	"net"
	"os"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/substrafoundation/substra-orchestrator/database/couchdb"
	"github.com/substrafoundation/substra-orchestrator/lib/assets/node"
	"github.com/substrafoundation/substra-orchestrator/lib/assets/objective"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// RunServerWithChainCode is exported
func RunServerWithChainCode() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatal("failed guess where")
	}

	if !wallet.Exists("appClient") {
		cert, err := ioutil.ReadFile("/Users/inal/fabric/sampleconfig/msp/signcerts/peer.pem")
		if err != nil {
			log.Fatal("failed guess where")
		}

		key, err := ioutil.ReadFile("/Users/inal/fabric/sampleconfig/msp/keystore/key.pem")
		if err != nil {
			log.Fatal("failed guess where")
		}

		identity := gateway.NewX509Identity("SampleOrg.member", string(cert), string(key))

		wallet.Put("appClient", identity)
	}

	// get config path

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile("/Users/inal/fabric/sampleconfig/config.json")),
		gateway.WithIdentity(wallet, "appClient"),
	)

	if err != nil {
		log.Fatalf("failed guess where %v", err)
	}

	defer gw.Close()

	network, err := gw.GetNetwork("ch1")
	if err != nil {
		log.Fatal("failed guess where")
	}

	contract := network.GetContract("mycc")
	result, err := contract.SubmitTransaction("RegisterNode", "1")
	if err != nil {
		log.Fatal("failed guess where")
	}

	log.Debug(result)
}

// RunServerWithoutChainCode will expose the chaincode logic through gRPC.
// State will be stored in a redis database.
func RunServerWithoutChainCode() {
	dsn := "http://dev:dev@localhost:5984"

	couchPersistence, err := couchdb.NewPersistence(context.TODO(), dsn, "substra_orchestrator")
	defer couchPersistence.Close(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	server := grpc.NewServer()
	node.RegisterNodeServiceServer(server, node.NewServer(node.NewService(couchPersistence)))
	objective.RegisterObjectiveServiceServer(server, objective.NewServer(objective.NewService(couchPersistence)))

	reflection.Register(server)

	log.WithField("address", listen.Addr().String()).Info("Server listening")
	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to server grpc server on port 9000: %v", err)
	}
}

func main() {
	cLog := console.New(true)
	log.AddHandler(cLog, log.AllLevels...)

	RunServerWithoutChainCode()
}
