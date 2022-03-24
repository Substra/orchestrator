package scenarios

import (
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
)

var eventTestScenarios = []Scenario{
	{
		testEventTSFilter,
		[]string{"short", "event"},
	},
}

// testEventTSFilter will register some assets to generate events and filter event by timestamp.
func testEventTSFilter(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	nbTasks := 10

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	newTasks := make([]client.Taskable, 0, nbTasks)
	for i := 0; i < nbTasks; i++ {
		newTasks = append(newTasks, client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(client.DefaultTaskRef))
	}
	appClient.RegisterTasks(newTasks...)

	events := appClient.QueryEvents(&asset.EventQueryFilter{}, "", 100)

	bound := events.Events[5].Timestamp
	eventsBetween := appClient.QueryEvents(&asset.EventQueryFilter{Start: bound, End: bound}, "", 100)
	for _, e := range eventsBetween.Events {
		if !e.Timestamp.AsTime().Equal(bound.AsTime()) {
			log.Fatalf("Unexpected value for event timestamp. Expected %s, got %s", bound.AsTime(), e.Timestamp.AsTime())
		}
	}

	eventsBefore := appClient.QueryEvents(&asset.EventQueryFilter{End: bound}, "", 100)
	for _, e := range eventsBefore.Events {
		if e.Timestamp.AsTime().After(bound.AsTime()) {
			log.Fatalf("Unexpected value for event timestamp. Expected a value lower than %s, got %s", bound.AsTime(), e.Timestamp.AsTime())
		}
	}

	eventsAfter := appClient.QueryEvents(&asset.EventQueryFilter{Start: bound}, "", 100)
	for _, e := range eventsAfter.Events {
		if e.Timestamp.AsTime().Before(bound.AsTime()) {
			log.Fatalf("Unexpected value for event timestamp. Expected a value greater than %s, got %s", bound.AsTime(), e.Timestamp.AsTime())
		}
	}

	allEvents := appClient.QueryEvents(&asset.EventQueryFilter{Start: events.Events[0].Timestamp, End: events.Events[len(events.Events)-1].Timestamp}, "", 100)
	if len(allEvents.Events) != len(events.Events) {
		log.Fatalf("Unexpected number of events. Expected %d, got %d", len(events.Events), len(allEvents.Events))
	}

}
