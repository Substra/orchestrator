//go:build e2e
// +build e2e

package e2e

import (
	"github.com/owkin/orchestrator/e2e/client"
	"google.golang.org/grpc"
)

var factory *client.TestClientFactory

func initTestClientFactory(conn *grpc.ClientConn, mspid, channel, chaincode string) {
	factory = client.NewTestClientFactory(conn, mspid, channel, chaincode)
}
