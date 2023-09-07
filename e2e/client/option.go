//go:build e2e
// +build e2e

package client

import (
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
)

type ComputePlanOptions struct {
	KeyRef string
}

type FunctionOptions struct {
	KeyRef  string
	Inputs  map[string]*asset.FunctionInput
	Outputs map[string]*asset.FunctionOutput
}

type DataSampleOptions struct {
	KeyRef string
}

type TaskOutputRef struct {
	TaskRef    string
	Identifier string
}

type TaskInputOptions struct {
	Identifier string
	TaskOutput *TaskOutputRef // either this field is set
	AssetRef   string         //...or this field is set
}

type TrainTaskOptions struct {
	KeyRef         string
	FunctionRef    string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type TestTaskOptions struct {
	KeyRef         string
	FunctionRef    string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type PredictTaskOptions struct {
	KeyRef         string
	FunctionRef    string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type CompositeTaskOptions struct {
	KeyRef         string
	FunctionRef    string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type AggregateTaskOptions struct {
	KeyRef      string
	FunctionRef string
	PlanRef     string
	Worker      string
	Inputs      []*TaskInputOptions
	Outputs     map[string]*asset.NewComputeTaskOutput
}

type ModelOptions struct {
	KeyRef     string
	TaskRef    string
	TaskOutput string
}

type PerformanceOptions struct {
	ComputeTaskKeyRef string
	ComputeTaskOutput string
	MetricKeyRef      string
	PerformanceValue  float32
}

type DataManagerOptions struct {
	LogsPermission *asset.NewPermissions
}

func DefaultTestTaskOptions() *TestTaskOptions {
	return &TestTaskOptions{
		KeyRef:         DefaultTestTaskRef,
		FunctionRef:    DefaultMetricFunctionRef,
		PlanRef:        DefaultPlanRef,
		DataManagerRef: DefaultDataManagerRef,
		DataSampleRef:  DefaultDataSampleRef,
		Inputs: []*TaskInputOptions{
			{
				Identifier: "opener",
				AssetRef:   DefaultDataManagerRef,
			},
			{
				Identifier: "datasamples",
				AssetRef:   DefaultDataSampleRef,
			},
		},
		Outputs: map[string]*asset.NewComputeTaskOutput{
			"performance": {
				Permissions: &asset.NewPermissions{Public: true},
			},
		},
	}
}

func (o *TestTaskOptions) WithDataSampleRef(ref string) *TestTaskOptions {
	o.DataSampleRef = ref
	return o
}

func (o *TestTaskOptions) WithKeyRef(ref string) *TestTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *TestTaskOptions) WithFunctionRef(ref string) *TestTaskOptions {
	o.FunctionRef = ref
	return o
}

func (o *TestTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		FunctionKey:    ks.GetKey(o.FunctionRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Inputs:         GetNewTaskInputs(ks, o.Inputs),
		Outputs:        o.Outputs,
	}
}

func DefaultTrainTaskOptions() *TrainTaskOptions {
	return &TrainTaskOptions{
		KeyRef:         DefaultTrainTaskRef,
		FunctionRef:    DefaultSimpleFunctionRef,
		PlanRef:        DefaultPlanRef,
		DataManagerRef: DefaultDataManagerRef,
		DataSampleRef:  DefaultDataSampleRef,
		Inputs: []*TaskInputOptions{
			{
				Identifier: "opener",
				AssetRef:   DefaultDataManagerRef,
			},
			{
				Identifier: "datasamples",
				AssetRef:   DefaultDataSampleRef,
			},
		},
		Outputs: map[string]*asset.NewComputeTaskOutput{
			"model": {Permissions: &asset.NewPermissions{Public: true}, Transient: false},
		},
	}
}

func (o *TrainTaskOptions) WithKeyRef(ref string) *TrainTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *TrainTaskOptions) WithPlanRef(ref string) *TrainTaskOptions {
	o.PlanRef = ref
	return o
}

func (o *TrainTaskOptions) WithFunctionRef(ref string) *TrainTaskOptions {
	o.FunctionRef = ref
	return o
}

func (o *TrainTaskOptions) WithOutput(identifier string, permissions *asset.NewPermissions, transient bool) *TrainTaskOptions {
	o.Outputs[identifier] = &asset.NewComputeTaskOutput{
		Permissions: permissions,
		Transient:   transient,
	}
	return o
}

// SetOutputs will override existing outputs with provided argument
func (o *TrainTaskOptions) SetOutputs(outputs map[string]*asset.NewComputeTaskOutput) *TrainTaskOptions {
	o.Outputs = outputs
	return o
}

func (o *TrainTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		FunctionKey:    ks.GetKey(o.FunctionRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Inputs:         GetNewTaskInputs(ks, o.Inputs),
		Outputs:        o.Outputs,
	}
}

func DefaultPredictTaskOptions() *PredictTaskOptions {
	return &PredictTaskOptions{
		KeyRef:         DefaultPredictTaskRef,
		FunctionRef:    DefaultPredictFunctionRef,
		PlanRef:        DefaultPlanRef,
		DataManagerRef: DefaultDataManagerRef,
		DataSampleRef:  DefaultDataSampleRef,
		Inputs: []*TaskInputOptions{
			{
				Identifier: "opener",
				AssetRef:   DefaultDataManagerRef,
			},
			{
				Identifier: "datasamples",
				AssetRef:   DefaultDataSampleRef,
			},
		},
		Outputs: map[string]*asset.NewComputeTaskOutput{
			"predictions": {Permissions: &asset.NewPermissions{Public: false}},
		},
	}
}

func (o *PredictTaskOptions) WithFunctionRef(ref string) *PredictTaskOptions {
	o.FunctionRef = ref
	return o
}

func (o *PredictTaskOptions) WithKeyRef(ref string) *PredictTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *PredictTaskOptions) WithDataSampleRef(ref string) *PredictTaskOptions {
	o.DataSampleRef = ref
	return o
}

func (o *PredictTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		FunctionKey:    ks.GetKey(o.FunctionRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Inputs:         GetNewTaskInputs(ks, o.Inputs),
		Outputs:        o.Outputs,
	}
}

func DefaultCompositeTaskOptions() *CompositeTaskOptions {
	return &CompositeTaskOptions{
		KeyRef:         DefaultCompositeTaskRef,
		FunctionRef:    DefaultCompositeFunctionRef,
		PlanRef:        DefaultPlanRef,
		DataManagerRef: DefaultDataManagerRef,
		DataSampleRef:  DefaultDataSampleRef,
		Inputs: []*TaskInputOptions{
			{
				Identifier: "opener",
				AssetRef:   DefaultDataManagerRef,
			},
			{
				Identifier: "datasamples",
				AssetRef:   DefaultDataSampleRef,
			},
		},
		Outputs: map[string]*asset.NewComputeTaskOutput{
			"shared": {Permissions: &asset.NewPermissions{Public: true}},
			"local":  {Permissions: &asset.NewPermissions{Public: true}},
		},
	}
}

func (o *CompositeTaskOptions) WithKeyRef(ref string) *CompositeTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *CompositeTaskOptions) WithFunctionRef(ref string) *CompositeTaskOptions {
	o.FunctionRef = ref
	return o
}

func (o *CompositeTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		FunctionKey:    ks.GetKey(o.FunctionRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Inputs:         GetNewTaskInputs(ks, o.Inputs),
		Outputs:        o.Outputs,
	}
}

func DefaultAggregateTaskOptions() *AggregateTaskOptions {
	return &AggregateTaskOptions{
		KeyRef:      DefaultAggregateTaskRef,
		FunctionRef: DefaultAggregateFunctionRef,
		PlanRef:     DefaultPlanRef,
		Worker:      "MyOrg1MSP",
		Inputs:      []*TaskInputOptions{},
		Outputs: map[string]*asset.NewComputeTaskOutput{
			"model": {Permissions: &asset.NewPermissions{Public: true}},
		},
	}
}

func (o *AggregateTaskOptions) WithKeyRef(ref string) *AggregateTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *AggregateTaskOptions) WithWorker(w string) *AggregateTaskOptions {
	o.Worker = w
	return o
}

func (o *AggregateTaskOptions) WithFunctionRef(ref string) *AggregateTaskOptions {
	o.FunctionRef = ref
	return o
}

func (o *AggregateTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		FunctionKey:    ks.GetKey(o.FunctionRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Worker:         o.Worker,
		Inputs:         GetNewTaskInputs(ks, o.Inputs),
		Outputs:        o.Outputs,
	}
}

func DefaultSimpleFunctionOptions() *FunctionOptions {
	return &FunctionOptions{
		KeyRef: DefaultSimpleFunctionRef,
		Inputs: map[string]*asset.FunctionInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"model":       {Kind: asset.AssetKind_ASSET_MODEL, Multiple: true, Optional: true},
		},
		Outputs: map[string]*asset.FunctionOutput{
			"model": {Kind: asset.AssetKind_ASSET_MODEL},
		},
		Status: asset.FunctionStatus_FUNCTION_STATUS_CREATED,
	}
}

func DefaultCompositeFunctionOptions() *FunctionOptions {
	return &FunctionOptions{
		KeyRef: DefaultCompositeFunctionRef,
		Inputs: map[string]*asset.FunctionInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"shared":      {Kind: asset.AssetKind_ASSET_MODEL, Optional: true},
			"local":       {Kind: asset.AssetKind_ASSET_MODEL, Optional: true},
		},
		Outputs: map[string]*asset.FunctionOutput{
			"shared": {Kind: asset.AssetKind_ASSET_MODEL},
			"local":  {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
}

func DefaultAggregateFunctionOptions() *FunctionOptions {
	return &FunctionOptions{
		KeyRef: DefaultAggregateFunctionRef,
		Inputs: map[string]*asset.FunctionInput{
			"model": {Kind: asset.AssetKind_ASSET_MODEL, Multiple: true, Optional: true},
		},
		Outputs: map[string]*asset.FunctionOutput{
			"model": {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
}

func DefaultPredictFunctionOptions() *FunctionOptions {
	return &FunctionOptions{
		KeyRef: DefaultPredictFunctionRef,
		Inputs: map[string]*asset.FunctionInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"model":       {Kind: asset.AssetKind_ASSET_MODEL},
		},
		Outputs: map[string]*asset.FunctionOutput{
			"predictions": {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
}

func DefaultMetricFunctionOptions() *FunctionOptions {
	return &FunctionOptions{
		KeyRef: DefaultMetricFunctionRef,
		Inputs: map[string]*asset.FunctionInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"predictions": {Kind: asset.AssetKind_ASSET_MODEL},
		},
		Outputs: map[string]*asset.FunctionOutput{
			"performance": {Kind: asset.AssetKind_ASSET_PERFORMANCE},
		},
	}
}

func (o *FunctionOptions) WithOutput(identifier string, kind asset.AssetKind, multiple bool) *FunctionOptions {
	o.Outputs[identifier] = &asset.FunctionOutput{
		Kind:     kind,
		Multiple: multiple,
	}
	return o
}

// SetInputs will override existing inputs with provided argument
func (o *FunctionOptions) SetInputs(inputs map[string]*asset.FunctionInput) *FunctionOptions {
	o.Inputs = inputs
	return o
}

// SetOutputs will override existing outputs with provided argument
func (o *FunctionOptions) SetOutputs(outputs map[string]*asset.FunctionOutput) *FunctionOptions {
	o.Outputs = outputs
	return o
}

func (o *FunctionOptions) WithKeyRef(ref string) *FunctionOptions {
	o.KeyRef = ref
	return o
}

func DefaultComputePlanOptions() *ComputePlanOptions {
	return &ComputePlanOptions{
		KeyRef: DefaultPlanRef,
	}
}

func (o *ComputePlanOptions) WithKeyRef(ref string) *ComputePlanOptions {
	o.KeyRef = ref
	return o
}

func DefaultModelOptions() *ModelOptions {
	return &ModelOptions{
		KeyRef:     DefaultModelRef,
		TaskRef:    DefaultTrainTaskRef,
		TaskOutput: "model",
	}
}

func (o *ModelOptions) WithKeyRef(ref string) *ModelOptions {
	o.KeyRef = ref
	return o
}

func (o *ModelOptions) WithTaskRef(ref string) *ModelOptions {
	o.TaskRef = ref
	return o
}

func (o *ModelOptions) WithTaskOutput(output string) *ModelOptions {
	o.TaskOutput = output
	return o
}

func DefaultPerformanceOptions() *PerformanceOptions {
	return &PerformanceOptions{
		ComputeTaskKeyRef: DefaultTrainTaskRef,
		ComputeTaskOutput: "performance",
		MetricKeyRef:      DefaultMetricFunctionRef,
		PerformanceValue:  0.5,
	}
}

func (o *PerformanceOptions) WithTaskRef(ref string) *PerformanceOptions {
	o.ComputeTaskKeyRef = ref
	return o
}

func (o *PerformanceOptions) WithTaskOutput(output string) *PerformanceOptions {
	o.ComputeTaskOutput = output
	return o
}

func DefaultDataSampleOptions() *DataSampleOptions {
	return &DataSampleOptions{
		KeyRef: "ds",
	}
}

func (o *DataSampleOptions) WithKeyRef(ref string) *DataSampleOptions {
	o.KeyRef = ref
	return o
}

func DefaultDataManagerOptions() *DataManagerOptions {
	return &DataManagerOptions{
		LogsPermission: &asset.NewPermissions{Public: true},
	}
}

func (o *DataManagerOptions) WithLogsPermission(permission *asset.NewPermissions) *DataManagerOptions {
	o.LogsPermission = permission
	return o
}

func GetNewTaskInputs(ks *KeyStore, inputs []*TaskInputOptions) []*asset.ComputeTaskInput {
	res := make([]*asset.ComputeTaskInput, len(inputs))

	for i, in := range inputs {
		res[i] = &asset.ComputeTaskInput{
			Identifier: in.Identifier,
		}

		if (in.AssetRef != "") && (in.TaskOutput != nil) {
			log.Fatal().Msg("Cannot have AssetRef and TaskOutput at the same time.")
		}

		if in.AssetRef != "" {
			res[i].Ref = &asset.ComputeTaskInput_AssetKey{
				AssetKey: ks.GetKey(in.AssetRef),
			}
		} else {
			res[i].Ref = &asset.ComputeTaskInput_ParentTaskOutput{
				ParentTaskOutput: &asset.ParentTaskOutputRef{
					ParentTaskKey:    ks.GetKey(in.TaskOutput.TaskRef),
					OutputIdentifier: in.TaskOutput.Identifier,
				},
			}
		}
	}

	return res
}
