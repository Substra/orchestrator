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
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/distributed"
	"github.com/owkin/orchestrator/server/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func getDistributedServer(tlsOption []grpc.ServerOption) common.Runnable {
	networkConfig := common.MustGetEnv("NETWORK_CONFIG")
	certificate := common.MustGetEnv("FABRIC_CERT")
	key := common.MustGetEnv("FABRIC_KEY")

	server, err := distributed.GetServer(networkConfig, certificate, key, tlsOption)
	if err != nil {
		log.WithError(err).Fatal("failed to create standalone server")
	}

	return server
}

func getStandaloneServer(tlsOptions []grpc.ServerOption) common.Runnable {
	dbURL := common.MustGetEnv("DATABASE_URL")
	rabbitDSN := common.MustGetEnv("AMQP_DSN")

	server, err := standalone.GetServer(dbURL, rabbitDSN, tlsOptions)
	if err != nil {
		log.WithError(err).Fatal("failed to create standalone server")
	}

	return server
}

func main() {
	var standaloneMode = false

	common.InitLogging()

	flag.BoolVar(&standaloneMode, "standalone", true, "Run the server in standalone mode")
	flag.BoolVar(&standaloneMode, "s", true, "Run the server in standalone mode (shorthand)")

	flag.Parse()

	mode, ok := common.GetEnv("MODE")
	if ok {
		switch mode {
		case "distributed":
			standaloneMode = false
		case "standalone":
			standaloneMode = true
		}
	}

	serverOptions := []grpc.ServerOption{}
	if tlsOptions := getTLSOptions(); tlsOptions != nil {
		serverOptions = append(serverOptions, tlsOptions)
	}

	var app common.Runnable
	if standaloneMode {
		app = getStandaloneServer(serverOptions)
	} else {
		app = getDistributedServer(serverOptions)
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
