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
	"os"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/distributed"
	"github.com/owkin/orchestrator/server/standalone"
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

func getDistributedServer() common.Runnable {
	networkConfig := mustGetEnv("NETWORK_CONFIG")
	certificate := mustGetEnv("CERT")
	key := mustGetEnv("KEY")

	server, err := distributed.GetServer(networkConfig, certificate, key)
	if err != nil {
		log.WithError(err).Fatal("failed to create standalone server")
	}

	return server
}

func getStandaloneServer() common.Runnable {
	dbURL := mustGetEnv("DATABASE_URL")
	rabbitDSN := mustGetEnv("AMQP_DSN")

	server, err := standalone.GetServer(dbURL, rabbitDSN)
	if err != nil {
		log.WithError(err).Fatal("failed to create standalone server")
	}

	return server
}

func main() {
	common.InitLogging()

	flag.BoolVar(&standaloneMode, "standalone", true, "Run the chaincode in standalone mode")
	flag.BoolVar(&standaloneMode, "s", true, "Run the chaincode in standalone mode (shorthand)")

	flag.Parse()

	switch os.Getenv(envPrefix + "MODE") {
	case "chaincode":
		standaloneMode = false
	case "standalone":
		standaloneMode = true
	}

	var app common.Runnable
	if standaloneMode {
		app = getStandaloneServer()
	} else {
		app = getDistributedServer()
	}
	defer app.Stop()

	// Register reflection service
	reflection.Register(app.GetGrpcServer())

	// Register healthcheck service
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(app.GetGrpcServer(), healthcheck)

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.WithError(err).Fatal("failed to listen on port 9000")
	}

	log.WithField("address", listen.Addr().String()).Info("Server listening")
	if err := app.GetGrpcServer().Serve(listen); err != nil {
		log.WithError(err).Fatal("failed to server grpc server on port 9000")
	}
}
