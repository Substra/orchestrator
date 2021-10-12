package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewComputeTaskValidation(t *testing.T) {
	validTrainTask := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		Category:       ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Data: &NewComputeTask_Train{
			Train: &NewTrainTaskData{
				DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
				DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
			},
		},
	}
	invalidCategory := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		Category:       ComputeTaskCategory_TASK_UNKNOWN,
		AlgoKey:        "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Metadata:       map[string]string{"test": "indeed"},
		ParentTaskKeys: []string{"7ae86bc1-aa4a-492f-90a6-ad5e686afb8f"},
		Data: &NewComputeTask_Train{
			Train: &NewTrainTaskData{
				DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
				DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
			},
		},
	}
	missingAlgo := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		Category:       ComputeTaskCategory_TASK_TRAIN,
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Metadata:       map[string]string{"test": "indeed"},
		ParentTaskKeys: []string{"7ae86bc1-aa4a-492f-90a6-ad5e686afb8f"},
		Data: &NewComputeTask_Train{
			Train: &NewTrainTaskData{
				DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
				DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
			},
		},
	}
	missingComputePlan := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		Category:       ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        "867852b4-8419-4d52-8862-d5db823095be",
		Metadata:       map[string]string{"test": "indeed"},
		ParentTaskKeys: []string{"7ae86bc1-aa4a-492f-90a6-ad5e686afb8f"},
		Data: &NewComputeTask_Train{
			Train: &NewTrainTaskData{
				DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
				DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
			},
		},
	}
	missingData := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		Category:       ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Metadata:       map[string]string{"test": "indeed"},
		ParentTaskKeys: []string{"7ae86bc1-aa4a-492f-90a6-ad5e686afb8f"},
	}

	cases := map[string]struct {
		valid   bool
		newTask *NewComputeTask
	}{
		"valid":                {valid: true, newTask: validTrainTask},
		"invalid category":     {valid: false, newTask: invalidCategory},
		"missing algokey":      {valid: false, newTask: missingAlgo},
		"missing compute plan": {valid: false, newTask: missingComputePlan},
		"missing train data":   {valid: false, newTask: missingData},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.newTask.Validate())
			} else {
				assert.Error(t, c.newTask.Validate())
			}
		})
	}
}

func TestNewTrainTaskDataValidation(t *testing.T) {
	valid := &NewTrainTaskData{
		DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
	}
	invalidSampleKey := &NewTrainTaskData{
		DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataSampleKeys: []string{"not a uuid", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
	}
	invalidManagerKey := &NewTrainTaskData{
		DataManagerKey: "not a uuid",
		DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
	}

	cases := map[string]struct {
		valid bool
		data  *NewTrainTaskData
	}{
		"valid":           {valid: true, data: valid},
		"invalid manager": {valid: false, data: invalidManagerKey},
		"invalid samples": {valid: false, data: invalidSampleKey},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.data.Validate())
			} else {
				assert.Error(t, c.data.Validate())
			}
		})
	}
}

func TestNewTestTaskDataValidation(t *testing.T) {
	validDataSamples := &NewTestTaskData{
		MetricKey:      "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
	}
	missingMetric := &NewTestTaskData{
		DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
	}
	missingSamples := &NewTestTaskData{
		MetricKey:      "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
	}

	cases := map[string]struct {
		valid bool
		data  *NewTestTaskData
	}{
		"with samples":    {valid: true, data: validDataSamples},
		"missing metrics": {valid: false, data: missingMetric},
		"missing samples": {valid: false, data: missingSamples},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.data.Validate())
			} else {
				assert.Error(t, c.data.Validate())
			}
		})
	}
}

func TestNewAggregateTrainTaskDataValidation(t *testing.T) {
	valid := &NewAggregateTrainTaskData{
		Worker: "MyORG2MSP",
	}
	empty := &NewAggregateTrainTaskData{}

	cases := map[string]struct {
		valid bool
		data  *NewAggregateTrainTaskData
	}{
		"valid": {valid: true, data: valid},
		"empty": {valid: false, data: empty},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.data.Validate())
			} else {
				assert.Error(t, c.data.Validate())
			}
		})
	}
}

func TestNewCompositeTrainTaskDataValidation(t *testing.T) {
	perms := &NewPermissions{
		Public: true,
	}
	valid := &NewCompositeTrainTaskData{
		DataManagerKey:   "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataSampleKeys:   []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
		TrunkPermissions: perms,
	}
	missingPerms := &NewCompositeTrainTaskData{
		DataManagerKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
		DataSampleKeys: []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
	}
	invalidManager := &NewCompositeTrainTaskData{
		DataManagerKey:   "not a uuid",
		DataSampleKeys:   []string{"85e39014-ae2e-4fa4-b05b-4437076a4fa7", "8a90a6e3-2e7e-4c9d-9ed3-47b99942d0a8"},
		TrunkPermissions: perms,
	}

	cases := map[string]struct {
		valid bool
		data  *NewCompositeTrainTaskData
	}{
		"valid":           {valid: true, data: valid},
		"missing perms":   {valid: false, data: missingPerms},
		"invalid manager": {valid: false, data: invalidManager},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.data.Validate())
			} else {
				assert.Error(t, c.data.Validate())
			}
		})
	}
}

func TestApplyTaskActionParam(t *testing.T) {
	empty := &ApplyTaskActionParam{}
	valid := &ApplyTaskActionParam{
		ComputeTaskKey: "972bef4c-1b42-4743-bbe9-cc3f4a69952f",
		Action:         ComputeTaskAction_TASK_ACTION_DOING,
	}
	missingKey := &ApplyTaskActionParam{
		Action: ComputeTaskAction_TASK_ACTION_DOING,
	}
	missingAction := &ApplyTaskActionParam{
		ComputeTaskKey: "972bef4c-1b42-4743-bbe9-cc3f4a69952f",
	}
	invalidAction := &ApplyTaskActionParam{
		ComputeTaskKey: "972bef4c-1b42-4743-bbe9-cc3f4a69952f",
		Action:         ComputeTaskAction_TASK_ACTION_UNKNOWN,
	}

	cases := map[string]struct {
		valid bool
		param *ApplyTaskActionParam
	}{
		"valid":          {valid: true, param: valid},
		"empty":          {valid: false, param: empty},
		"missing key":    {valid: false, param: missingKey},
		"missing action": {valid: false, param: missingAction},
		"invalid action": {valid: false, param: invalidAction},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.valid {
				assert.NoError(t, c.param.Validate())
			} else {
				assert.Error(t, c.param.Validate())
			}
		})
	}
}
