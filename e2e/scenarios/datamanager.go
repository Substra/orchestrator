package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/golang/protobuf/proto"
	"github.com/owkin/orchestrator/e2e/client"
)

var datamanagerTestScenarios = []Scenario{
	{
		testRegisterDataManager,
		[]string{"short", "datamanager"},
	},
}

func testRegisterDataManager(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	registeredManager := appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	retrievedManager := appClient.GetDataManager(client.DefaultDataManagerRef)

	if !proto.Equal(registeredManager, retrievedManager) {
		log.WithField("registeredManager", registeredManager).WithField("retrievedManager", retrievedManager).
			Fatal("The retrieved datamanager differs from the registered datamanager")
	}
}
