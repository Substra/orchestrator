//go:build e2e
// +build e2e

package e2e

import (
	"github.com/substra/orchestrator/e2e/client"
	"google.golang.org/grpc"
)

var factory *client.TestClientFactory

func initTestClientFactory(conn *grpc.ClientConn, mspid, channel) {
	factory = client.NewTestClientFactory(conn, mspid, channel)
}
