package scenarios

import (
	"strings"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/e2e/client"
	"github.com/owkin/orchestrator/lib/asset"
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
}

// register a test task, start it, and register its performance.
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

	_, err := appClient.RegisterPerformance(client.DefaultPerformanceOptions().WithTaskRef("testTask").WithMetricRef("testmetric"))
	if err != nil {
		log.WithError(err).Fatal("RegisterPerformance failed")
	}
	task := appClient.GetComputeTask("testTask")
	if task.Status != asset.ComputeTaskStatus_STATUS_DONE {
		log.Fatal("test task should be DONE")
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
