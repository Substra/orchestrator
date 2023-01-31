//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
	e2erequire "github.com/substra/orchestrator/e2e/require"
	"github.com/substra/orchestrator/lib/asset"
	"golang.org/x/sync/errgroup"
)

// TestEventTSFilter will register some assets to generate events and filter event by timestamp.
func TestEventTSFilter(t *testing.T) {
	appClient := factory.NewTestClient()

	nbTasks := 10

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())

	newTasks := make([]client.Taskable, 0, nbTasks)
	for i := 0; i < nbTasks; i++ {
		newTasks = append(newTasks, client.DefaultTrainTaskOptions().
			WithKeyRef(fmt.Sprintf("task%d", i)).
			WithInput("model", &client.TaskOutputRef{TaskRef: client.DefaultTrainTaskRef, Identifier: "model"}))
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

// TestSubscribeReplayEvents ensures that previous events can be replayed.
func TestSubscribeReplayEvents(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	plan := appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	newTasks := make([]client.Taskable, 20)
	for i := range newTasks {
		newTasks[i] = client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i))
	}
	tasks := appClient.RegisterTasks(newTasks...)

	startEvent := appClient.GetAssetCreationEvent(plan.Key)
	stream, cancel := appClient.SubscribeToEvents(startEvent.Id)
	defer cancel()

	for _, task := range tasks {
		event, err := stream.Recv()
		require.NoError(t, err)
		e2erequire.ProtoEqual(t, task, event.GetComputeTask())
	}
}

// TestSubscribeEventsEmittedWhileSubscribed ensures that events emitted while
// the client is subscribed are forwarded.
func TestSubscribeEventsEmittedWhileSubscribed(t *testing.T) {
	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	plan := appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	startEvent := appClient.GetAssetCreationEvent(plan.Key)
	stream, cancel := appClient.SubscribeToEvents(startEvent.Id)
	defer cancel()

	newTasks := make([]client.Taskable, 20)
	for i := range newTasks {
		newTasks[i] = client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i))
	}
	tasks := appClient.RegisterTasks(newTasks...)

	for _, task := range tasks {
		event, err := stream.Recv()
		require.NoError(t, err)
		e2erequire.ProtoEqual(t, task, event.GetComputeTask())
	}
}

// TestSubscribeReplayThenListen ensures that after previous events have been replayed,
// it is possible to listen to new events.
func TestSubscribeReplayThenListen(t *testing.T) {
	appClient := factory.NewTestClient()
	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	manager := appClient.RegisterDataManager(client.DefaultDataManagerOptions())

	replayedSamples := make([]*asset.DataSample, 5)
	for i := range replayedSamples {
		replayedSamples[i] = appClient.RegisterDataSample(
			client.DefaultDataSampleOptions().WithKeyRef(fmt.Sprintf("replayedSample%d", i)),
		)
	}

	startEvent := appClient.GetAssetCreationEvent(manager.Key)
	stream, cancel := appClient.SubscribeToEvents(startEvent.Id)
	defer cancel()

	newSamples := make([]*asset.DataSample, 5)
	for i := range newSamples {
		newSamples[i] = appClient.RegisterDataSample(
			client.DefaultDataSampleOptions().WithKeyRef(fmt.Sprintf("newSample%d", i)),
		)
	}

	for _, sample := range append(replayedSamples, newSamples...) {
		event, err := stream.Recv()
		require.NoError(t, err)
		e2erequire.ProtoEqual(t, sample, event.GetDataSample())
	}
}

// TestSubscribeWithoutStartEventID ensures it is possible to subscribe
// to events without providing a startEventID param.
func TestSubscribeWithoutStartEventID(t *testing.T) {
	appClient := factory.NewTestClient()
	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())

	stream, cancel := appClient.SubscribeToEvents("")
	defer cancel()

	appClient.RegisterDataManager(client.DefaultDataManagerOptions())

	for i := 0; i < 2; i++ {
		event, err := stream.Recv()
		require.NoError(t, err)
		require.NotEqual(t, event.AssetKey, "")
	}
}

// TestSubscribeCheckEventStreamOrder ensures that events in the stream are
// ordered by timestamp and that they belong to the client channel.
// This check is made on a stream of events containing replayed events but also
// events that are emitted while listening.
func TestSubscribeCheckEventStreamConsistency(t *testing.T) {
	client1 := factory.WithChaincode("mycc").WithChannel("mychannel").NewTestClient()
	client2 := factory.WithChaincode("yourcc").WithChannel("yourchannel").NewTestClient()

	function := client1.RegisterFunction(client.DefaultSimpleFunctionOptions())
	client2.RegisterFunction(client.DefaultSimpleFunctionOptions())

	client1.RegisterDataManager(client.DefaultDataManagerOptions())
	client2.RegisterDataManager(client.DefaultDataManagerOptions())

	client1.RegisterDataSample(client.DefaultDataSampleOptions())
	client2.RegisterDataSample(client.DefaultDataSampleOptions())

	client1.RegisterComputePlan(client.DefaultComputePlanOptions())
	client2.RegisterComputePlan(client.DefaultComputePlanOptions())

	startEvent := client1.GetAssetCreationEvent(function.Key)
	stream, cancel := client1.SubscribeToEvents(startEvent.Id)
	defer cancel()

	nbTasks := 2
	// register tasks in separate requests to ensure that task creations
	// result in events with different timestamps
	for i := 0; i < nbTasks; i++ {
		taskOpt := client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i))
		client1.RegisterTasks(taskOpt)
		client2.RegisterTasks(taskOpt)
	}

	var previous *asset.Event
	for i := 0; i < 3+nbTasks; i++ {
		event, err := stream.Recv()
		require.NoError(t, err)

		require.Equal(t, event.Channel, client1.Channel)

		if previous != nil {
			require.GreaterOrEqual(t, event.Timestamp.AsTime(), previous.Timestamp.AsTime())
		}

		previous = event
	}
}

// TestSubscribeCheckEventStreamOrder ensures that two clients subscribed to the same the stream
// receive the same events in parallel.
func TestSubscribeParallel(t *testing.T) {
	client1 := factory.NewTestClient()
	client2 := factory.NewTestClient()

	client1.RegisterFunction(client.DefaultSimpleFunctionOptions())
	client1.RegisterDataManager(client.DefaultDataManagerOptions())
	client1.RegisterDataSample(client.DefaultDataSampleOptions())
	plan := client1.RegisterComputePlan(client.DefaultComputePlanOptions())

	replayedTasks := make([]client.Taskable, 101)
	for i := range replayedTasks {
		replayedTasks[i] = client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("replayedTask%d", i))
	}
	registeredTasks := client1.RegisterTasks(replayedTasks...)

	startEventId := client1.GetAssetCreationEvent(plan.Key).Id
	stream1, cancel1 := client1.SubscribeToEvents(startEventId)
	stream2, cancel2 := client2.SubscribeToEvents(startEventId)

	for range replayedTasks {
		event1, err := stream1.Recv()
		require.NoError(t, err)

		event2, err := stream2.Recv()
		require.NoError(t, err)

		e2erequire.ProtoEqual(t, event1, event2)
	}

	cancel1()
	cancel2()

	startEventId = client1.GetAssetCreationEvent(registeredTasks[len(registeredTasks)-1].Key).Id
	stream1, cancel1 = client1.SubscribeToEvents(startEventId)
	defer cancel1()
	stream2, cancel2 = client2.SubscribeToEvents(startEventId)
	defer cancel2()

	newTasks := make([]client.Taskable, 20)
	for i := range newTasks {
		newTasks[i] = client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("newTask%d", i))
	}
	client1.RegisterTasks(newTasks...)

	for range newTasks {
		event1, err := stream1.Recv()
		require.NoError(t, err)

		event2, err := stream2.Recv()
		require.NoError(t, err)

		e2erequire.ProtoEqual(t, event1, event2)
	}
}

// TestSubscribeSpeedSingleClient tests how long it takes to receive a given
// number of newly emitted events while a single client is subscribed.
func TestSubscribeSpeedSingleClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping speed test")
	}

	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	plan := appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	startEvent := appClient.GetAssetCreationEvent(plan.Key)
	stream, cancel := appClient.SubscribeToEvents(startEvent.Id)
	defer cancel()

	nbTasks := 1000
	newTasks := make([]client.Taskable, nbTasks)
	for i := range newTasks {
		newTasks[i] = client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i))
	}
	appClient.RegisterTasks(newTasks...)

	start := time.Now()
	for range newTasks {
		_, err := stream.Recv()
		require.NoError(t, err)
	}
	elapsed := time.Since(start)
	t.Logf("Received %d events in %s", nbTasks, elapsed)
}

// TestSubscribeSpeedMultipleClients tests how long it takes to receive a given
// number of newly emitted events while multiple concurrent clients are subscribed.
func TestSubscribeSpeedMultipleClients(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping speed test")
	}

	appClient := factory.NewTestClient()

	appClient.RegisterFunction(client.DefaultSimpleFunctionOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	plan := appClient.RegisterComputePlan(client.DefaultComputePlanOptions())

	startEvent := appClient.GetAssetCreationEvent(plan.Key)

	nbTasks := 1000
	newTasks := make([]client.Taskable, nbTasks+1)
	for i := range newTasks {
		newTasks[i] = client.DefaultTrainTaskOptions().WithKeyRef(fmt.Sprintf("task%d", i))
	}

	nbClients := 10
	g := new(errgroup.Group)
	for i := 0; i < nbClients; i++ {
		clientNum := i

		g.Go(func() error {
			stream, cancel := appClient.SubscribeToEvents(startEvent.Id)
			defer cancel()

			_, err := stream.Recv()
			if err != nil {
				return err
			}

			// start the timer after receiving the first event
			start := time.Now()
			for j := 0; j < nbTasks; j++ {
				_, err := stream.Recv()
				if err != nil {
					return err
				}
			}
			elapsed := time.Since(start)
			t.Logf("Received %d events with client %d in %s", nbTasks, clientNum, elapsed)

			return nil
		})
	}

	appClient.RegisterTasks(newTasks...)

	require.NoError(t, g.Wait())
}
