package scenarios

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/owkin/orchestrator/e2e/client"
)

type Scenario struct {
	Exec func(*client.TestClientFactory)
	Tags []string
}

var testScenarios = [][]Scenario{
	algoTestScenarios,
	computePlanTestScenarios,
	computeTaskTestScenarios,
	datasampleTestsScenarios,
	datasetTestScenarios,
	eventTestScenarios,
	failureReportScenarios,
	modelTestScenarios,
	performanceTestScenarios,
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func getScenarioName(s Scenario) string {
	funcName := getFunctionName(s.Exec)
	split := strings.Split(funcName, ".")
	return split[len(split)-1]
}

func GatherTestScenarios() map[string]Scenario {
	gatheredScenarios := make(map[string]Scenario)

	for _, scenarios := range testScenarios {
		for _, scenario := range scenarios {
			name := getScenarioName(scenario)
			gatheredScenarios[name] = scenario
		}
	}

	return gatheredScenarios
}
