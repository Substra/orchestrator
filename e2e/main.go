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

// Package main implements end to end testing client.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	debugEnabled = flag.Bool("debug", false, "Debug mode (very verbose)")
	tlsEnabled   = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile       = flag.String("cafile", "", "The file containing the CA root cert file")
	certFile     = flag.String("certfile", "", "The file containing the client cert file")
	keyFile      = flag.String("keyfile", "", "The file containing the client cert key")
	serverAddr   = flag.String("server_addr", "localhost:9000", "The server address in the format of host:port")
	mspid        = flag.String("mspid", "MyOrg1MSP", "MSP ID")
	channel      = flag.String("channel", "mychannel", "Channel to use")
	chaincode    = flag.String("chaincode", "mycc", "Chaincode to use (only relevant in distributed mode)")
	testFilter   = flag.String("name", "", "Filter test by name")
)

type scenario = func(*grpc.ClientConn)

var testScenarios = map[string]scenario{
	"TrainTaskLifecycle":    testTrainTaskLifecycle,
	"RegisterModel":         testRegisterModel,
	"CascadeCancel":         testCascadeCancel,
	"CascadeTodo":           testCascadeTodo,
	"CascadeFailure":        testCascadeFailure,
	"DeleteIntermediary":    testDeleteIntermediary,
	"MultiStageComputePlan": testMultiStageComputePlan,
	"QueryTasks":            testQueryTasks,
	"RegisterPerformance":   testRegisterPerformance,
	"CompositeParentChild":  testCompositeParentChild,
	"Concurrency":           testConcurrency,
}

func main() {
	flag.Parse()

	cLog := console.New(true)
	levels := make([]log.Level, 0)
	for _, lvl := range log.AllLevels {
		if !*debugEnabled && lvl == log.DebugLevel {
			continue
		}

		levels = append(levels, lvl)
	}
	log.AddHandler(cLog, levels...)

	var opts []grpc.DialOption
	if *tlsEnabled {
		b, err := ioutil.ReadFile(*caFile)
		if err != nil {
			log.WithError(err).Fatal("failed to read cacert")
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			log.Fatal("failed to append cert")
		}
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Fatal("failed to load client keypair")
		}
		config := &tls.Config{
			InsecureSkipVerify: false,
			Certificates:       []tls.Certificate{cert},
			RootCAs:            cp,
		}
		creds := grpc.WithTransportCredentials(credentials.NewTLS(config))

		opts = append(opts, creds)
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())

	log.WithField("server_addr", *serverAddr).Info("connecting to server")

	dialCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	conn, err := grpc.DialContext(dialCtx, *serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	log.Debug("Starting testing")

	for name, sc := range testScenarios {
		if *testFilter != "" && *testFilter != name {
			// skip non matching test
			continue
		}
		logger := log.WithField("name", name)
		logger.Debug("starting scenario")
		func() {
			defer logger.WithTrace().Info("test succeeded")
			sc(conn)
		}()

	}
}
