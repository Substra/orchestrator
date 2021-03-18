// +build e2e

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

// Package e2e contains end to end tests of the orchestrator.
// Note that due to the complexity of the distributed mode, tests only targets standalone orchestration.
package e2e

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type testApp struct {
	runnable common.Runnable
	listener *bufconn.Listener
}

func newTestApp() *testApp {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Missing DATABASE_URL env var")
	}

	rabbitDSN := os.Getenv("RABBITMQ_DSN")
	if rabbitDSN == "" {
		log.Fatal("Missing RABBITMQ_DSN")
	}

	server, err := standalone.GetServer(dbURL, rabbitDSN, nil)
	if err != nil {
		log.Fatalf("Cannot initialize test server: %s", err.Error())
	}

	return &testApp{
		runnable: server,
		listener: bufconn.Listen(bufSize),
	}
}

func (a *testApp) listen() {
	go func() {
		if err := a.runnable.GetGrpcServer().Serve(a.listener); err != nil {
			log.Fatalf("failed to listen: %s", err.Error())
		}
	}()
}

func (a *testApp) getDialer(context.Context, string) (net.Conn, error) {
	return a.listener.Dial()
}

var app *testApp

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		// Skip e2e testing in short mode
		os.Exit(0)
	}

	app = newTestApp()
	app.listen()

	ret := m.Run()

	app.runnable.Stop()
	os.Exit(ret)
}

func TestRegisterNode(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(app.getDialer), grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	ctx = metadata.AppendToOutgoingContext(ctx, "mspid", "MyOrg1MSP", "channel", "mychannel")

	client := asset.NewNodeServiceClient(conn)
	resp, err := client.RegisterNode(ctx, &asset.NodeRegistrationParam{})
	if err != nil {
		t.Errorf("RegisterNode failed: %v", err)
	}

	if resp.GetId() != "MyOrg1MSP" {
		t.Errorf("Unexpected node ID: %s", resp.GetId())
	}
}
