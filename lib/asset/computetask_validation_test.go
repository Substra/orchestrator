package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewComputeTaskValidation(t *testing.T) {
	validOutputs := map[string]*NewComputeTaskOutput{
		"model": {
			Permissions: &NewPermissions{
				Public: true,
			},
		},
	}

	validTrainTask := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*ComputeTaskInput{
			{
				Identifier: "model",
				Ref: &ComputeTaskInput_AssetKey{
					AssetKey: "867852b4-8419-4d52-8862-d5db823095be",
				},
			},
			{
				Identifier: "model2",
				Ref: &ComputeTaskInput_ParentTaskOutput{
					ParentTaskOutput: &ParentTaskOutputRef{
						ParentTaskKey:    "867852b4-8419-4d52-8862-d5db823095be",
						OutputIdentifier: "model",
					},
				},
			},
		},
		Outputs: validOutputs,
	}
	missingFunction := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Metadata:       map[string]string{"test": "indeed"},
	}
	missingComputePlan := &NewComputeTask{
		Key:         "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey: "867852b4-8419-4d52-8862-d5db823095be",
		Metadata:    map[string]string{"test": "indeed"},
	}
	invalidOutputPermissionsIdentifier := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Outputs: map[string]*NewComputeTaskOutput{
			"": {
				Permissions: &NewPermissions{
					Public: true,
				},
			},
		},
	}
	missingInputIdentifier := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*ComputeTaskInput{
			{
				Identifier: "",
				Ref: &ComputeTaskInput_AssetKey{
					AssetKey: "2837f0b7-cb0e-4a98-9df2-68c116f65ad6",
				},
			},
		},
	}
	missingInputRef := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs:         []*ComputeTaskInput{{Identifier: "model"}},
	}
	invalidInputRef := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*ComputeTaskInput{
			{
				Identifier: "model",
				Ref: &ComputeTaskInput_AssetKey{
					AssetKey: "abc",
				},
			},
		},
	}
	missingInputTaskOutputKey := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*ComputeTaskInput{
			{
				Identifier: "model",
				Ref: &ComputeTaskInput_ParentTaskOutput{
					ParentTaskOutput: &ParentTaskOutputRef{
						ParentTaskKey:    "",
						OutputIdentifier: "model",
					},
				},
			},
		},
	}
	invalidInputTaskOutputKey := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*ComputeTaskInput{
			{
				Identifier: "model",
				Ref: &ComputeTaskInput_ParentTaskOutput{
					ParentTaskOutput: &ParentTaskOutputRef{
						ParentTaskKey:    "abc",
						OutputIdentifier: "model",
					},
				},
			},
		},
	}
	missingInputTaskOutputIdentifier := &NewComputeTask{
		Key:            "867852b4-8419-4d52-8862-d5db823095be",
		FunctionKey:    "867852b4-8419-4d52-8862-d5db823095be",
		ComputePlanKey: "867852b4-8419-4d52-8862-d5db823095be",
		Inputs: []*ComputeTaskInput{
			{
				Identifier: "model",
				Ref: &ComputeTaskInput_ParentTaskOutput{
					ParentTaskOutput: &ParentTaskOutputRef{
						ParentTaskKey:    "867852b4-8419-4d52-8862-d5db823095be",
						OutputIdentifier: "",
					},
				},
			},
		},
	}

	cases := map[string]struct {
		valid   bool
		newTask *NewComputeTask
	}{
		"valid":                                 {valid: true, newTask: validTrainTask},
		"missing functionkey":                   {valid: false, newTask: missingFunction},
		"missing compute plan":                  {valid: false, newTask: missingComputePlan},
		"missing input identifier":              {valid: false, newTask: missingInputIdentifier},
		"missing input ref":                     {valid: false, newTask: missingInputRef},
		"invalid input ref":                     {valid: false, newTask: invalidInputRef},
		"missing input task output key":         {valid: false, newTask: missingInputTaskOutputKey},
		"invalid intput task output key":        {valid: false, newTask: invalidInputTaskOutputKey},
		"missing input task output identifier":  {valid: false, newTask: missingInputTaskOutputIdentifier},
		"invalid output permissions identifier": {valid: false, newTask: invalidOutputPermissionsIdentifier},
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

func TestApplyTaskActionParam(t *testing.T) {
	empty := &ApplyTaskActionParam{}
	valid := &ApplyTaskActionParam{
		ComputeTaskKey: "972bef4c-1b42-4743-bbe9-cc3f4a69952f",
		Action:         ComputeTaskAction_TASK_ACTION_EXECUTING,
	}
	missingKey := &ApplyTaskActionParam{
		Action: ComputeTaskAction_TASK_ACTION_EXECUTING,
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
