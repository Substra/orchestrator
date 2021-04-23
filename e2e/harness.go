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

package e2e

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// testApp wraps an orchestrator service (aka the "app") to be tested
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

// getDialer returns a dialer to be used in grpc connection
func (a *testApp) getDialer(context.Context, string) (net.Conn, error) {
	return a.listener.Dial()
}

// AppClient is a client for the tested app
type AppClient struct {
	ctx  context.Context
	conn *grpc.ClientConn
	keys map[string]string
}

func newAppClient(mspid, channel string) (*AppClient, error) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(app.getDialer), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "mspid", mspid, "channel", channel)

	return &AppClient{
		ctx:  ctx,
		conn: conn,
		keys: make(map[string]string),
	}, nil
}

func (c *AppClient) Close() {
	c.conn.Close()
}

// GetKey will create a UUID or return the previously generated one.
// This is useful when building relationships between entities.
func (c *AppClient) GetKey(id string) string {
	k, ok := c.keys[id]
	if !ok {
		k = uuid.New().String()
		c.keys[id] = k
	}

	return k
}
