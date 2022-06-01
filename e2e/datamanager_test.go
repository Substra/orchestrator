//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/owkin/orchestrator/e2e/client"
	e2erequire "github.com/owkin/orchestrator/e2e/require"
)

func TestRegisterDataManager(t *testing.T) {
	appClient := factory.NewTestClient()

	registeredManager := appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	retrievedManager := appClient.GetDataManager(client.DefaultDataManagerRef)

	e2erequire.ProtoEqual(t, registeredManager, retrievedManager)
}
