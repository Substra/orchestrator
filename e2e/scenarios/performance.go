package scenarios

import (
	"fmt"
	"strings"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/protobuf/proto"
)

var performanceTestScenarios = []Scenario{
	{
		testRegisterPerformance,
		[]string{"short", "perf"},
	},
	{
		testRegisterMultiplePerformances,
		[]string{"short", "perf"},
	},
	{
		testRegisterMultiplePerformancesForSameMetric,
		[]string{"short", "perf"},
	},
	{
		testQueryPerformances,
		[]string{"query", "perf"},
	},
}

// Register a test task, start it, register its performance,
// and ensure an event containing the performance is recorded.
func testRegisterPerformance(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithDataSampleRef("testds").WithParentsRef(client.DefaultTaskRef).WithMetricsRef("testmetric"))
	appClient.StartTask("testTask")

	registeredPerf, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"))
	if err != nil {
		log.WithError(err).Fatal("RegisterPerformance failed")
	}

	task := appClient.GetComputeTask("testTask")
	if task.Status != asset.ComputeTaskStatus_STATUS_DONE {
		log.Fatal("test task should be DONE")
	}

	perfResp := appClient.QueryPerformances(&asset.PerformanceQueryFilter{
		ComputeTaskKey: task.Key,
	}, "", 100)

	if len(perfResp.Performances) != 1 {
		log.Fatalf("Unexpected number of performances. Expected 1, got %d", len(perfResp.Performances))
	}

	retrievedPerf := perfResp.Performances[0]
	if !proto.Equal(registeredPerf, retrievedPerf) {
		log.WithField("registeredPerf", registeredPerf).WithField("retrievedPerf", retrievedPerf).
			Fatal("The retrieved performance differs from the registered performance")
	}

	eventResp := appClient.QueryEvents(&asset.EventQueryFilter{
		AssetKey:  registeredPerf.GetKey(),
		AssetKind: asset.AssetKind_ASSET_PERFORMANCE,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}, "", 100)

	if len(eventResp.Events) != 1 {
		log.Fatalf("Unexpected number of events. Expected 1, got %d", len(eventResp.Events))
	}

	eventPerf := eventResp.Events[0].GetPerformance()
	if !proto.Equal(registeredPerf, eventPerf) {
		log.WithField("registeredPerf", registeredPerf).WithField("eventPerf", eventPerf).
			Fatal("The performance in the event differs from the registered performance")
	}
}

func testRegisterMultiplePerformances(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric1"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric2"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithDataSampleRef("testds").WithParentsRef(client.DefaultTaskRef).WithMetricsRef("testmetric1", "testmetric2"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric1"))
	if err != nil {
		log.WithError(err).Fatal("RegisterPerformance failed")
	}
	task := appClient.GetComputeTask("testTask")
	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		log.Fatal("test task should be DOING")
	}

	_, err = appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric2"))
	if err != nil {
		log.WithError(err).Fatal("RegisterPerformance failed")
	}
	task = appClient.GetComputeTask("testTask")
	if task.Status != asset.ComputeTaskStatus_STATUS_DONE {
		log.Fatal("test task should be DONE")
	}
}

func testRegisterMultiplePerformancesForSameMetric(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())

	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric1"))
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric2"))
	appClient.RegisterTasks(client.DefaultTestTaskOptions().WithKeyRef("testTask").WithDataSampleRef("testds").WithParentsRef(client.DefaultTaskRef).WithMetricsRef("testmetric1", "testmetric2"))
	appClient.StartTask("testTask")

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric1"))
	if err != nil {
		log.WithError(err).Fatal("RegisterPerformance failed")
	}
	task := appClient.GetComputeTask("testTask")
	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		log.Fatal("test task should be DOING")
	}

	_, err = appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric1"))
	if err == nil {
		log.Fatal("RegisterPerformance should have failed.")
		if !strings.Contains(err.Error(), "OE0003") {
			log.WithError(err).Fatal("Unexpected error code")
		}
	}
	task = appClient.GetComputeTask("testTask")
	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		log.Fatal("test task should be DOING")
	}
}

func testQueryPerformances(factory *client.TestClientFactory) {
	appClient := factory.NewTestClient()

	appClient.RegisterAlgo(client.DefaultAlgoOptions())
	appClient.RegisterDataManager(client.DefaultDataManagerOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions())
	appClient.RegisterDataSample(client.DefaultDataSampleOptions().WithKeyRef("testds").WithTestOnly(true))
	appClient.RegisterComputePlan(client.DefaultComputePlanOptions())
	appClient.RegisterTasks(client.DefaultTrainTaskOptions())
	appClient.StartTask(client.DefaultTaskRef)
	appClient.RegisterModel(client.DefaultModelOptions())
	appClient.RegisterAlgo(client.DefaultAlgoOptions().WithCategory(asset.AlgoCategory_ALGO_METRIC).WithKeyRef("testmetric"))

	const nbPerformances = 2
	for i := 0; i < nbPerformances; i++ {
		testTaskRef := fmt.Sprint("testTask", i)
		appClient.RegisterTasks(client.DefaultTestTaskOptions().
			WithKeyRef(testTaskRef).
			WithDataSampleRef("testds").
			WithParentsRef(client.DefaultTaskRef).
			WithMetricsRef("testmetric"))
		appClient.StartTask(testTaskRef)

		_, err := appClient.RegisterPerformance(
			client.DefaultPerformanceOptions().WithTaskRef(testTaskRef).WithMetricRef("testmetric"),
		)
		if err != nil {
			log.WithError(err).Fatal("RegisterPerformance failed")
		}
	}

	res := appClient.QueryPerformances(nil, "", 10)
	performances := res.Performances

	if len(performances) != nbPerformances {
		log.WithField("performances", performances).
			Fatal(fmt.Sprintf("Expected %d performance items, got %d", nbPerformances, len(performances)))
	}

	if performances[0].CreationDate.AsTime().After(performances[1].CreationDate.AsTime()) {
		log.Fatal("Unexpected performance ordering")
	}
}
