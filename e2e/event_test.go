//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"testing"

	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/require"
)

// TestEventTSFilter will register some assets to generate events and filter event by timestamp.
func TestEventTSFilter(t *testing.T) {
	appClient := factory.NewTestClient()

	nbTasks := 10

	appClient.RegisterAlgo(client.DefaultSimpleAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	newTasks := make([]client.Taskable, 0, nbTasks)
	for i := 0; i < nbTasks; i++ {
		newTasks = append(newTasks, client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i)).WithParentsRef(client.DefaultTrainTaskRef))
	}
	appClient.RegisterTasks(newTasks...)

	events := appClient.QueryEvents(&asset.EventQueryFilter{}, "", 100)

	bound := events.Events[5].Timestamp
	eventsBetween := appClient.QueryEvents(&asset.EventQueryFilter{Start: bound, End: bound}, "", 100)
	for _, e := range eventsBetween.Events {
		require.Equal(t, bound.AsTime(), e.Timestamp.AsTime())
	}

	eventsBefore := appClient.QueryEvents(&asset.EventQueryFilter{End: bound}, "", 100)
	for _, e := range eventsBefore.Events {
		require.LessOrEqual(t, e.Timestamp.AsTime(), bound.AsTime())
	}

	eventsAfter := appClient.QueryEvents(&asset.EventQueryFilter{Start: bound}, "", 100)
	for _, e := range eventsAfter.Events {
		require.GreaterOrEqual(t, e.Timestamp.AsTime(), bound.AsTime())
	}

	allEvents := appClient.QueryEvents(&asset.EventQueryFilter{Start: events.Events[0].Timestamp, End: events.Events[len(events.Events)-1].Timestamp}, "", 100)
	require.Equal(t, len(allEvents.Events), len(events.Events))
}
