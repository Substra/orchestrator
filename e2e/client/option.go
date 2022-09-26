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

type AlgoOptions struct {
	KeyRef   string
	Category asset.AlgoCategory
	Inputs   map[string]*asset.AlgoInput
	Outputs  map[string]*asset.AlgoOutput
}

type DataSampleOptions struct {
	KeyRef   string
	TestOnly bool
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
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type TestTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type PredictTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type CompositeTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
	Outputs        map[string]*asset.NewComputeTaskOutput
}

type AggregateTaskOptions struct {
	KeyRef     string
	AlgoRef    string
	ParentsRef []string
	PlanRef    string
	Worker     string
	Inputs     []*TaskInputOptions
	Outputs    map[string]*asset.NewComputeTaskOutput
}

type ModelOptions struct {
	KeyRef     string
	TaskRef    string
	TaskOutput string
	Category   asset.ModelCategory
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
		AlgoRef:        DefaultMetricAlgoRef,
		ParentsRef:     []string{DefaultPredictTaskRef},
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

func (o *TestTaskOptions) WithAlgoRef(ref string) *TestTaskOptions {
	o.AlgoRef = ref
	return o
}

func (o *TestTaskOptions) WithParentsRef(p ...string) *TestTaskOptions {
	o.ParentsRef = p
	return o
}

func (o *TestTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = ks.GetKey(ref)
	}
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_TEST,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Test{
			Test: &asset.NewTestTaskData{
				DataManagerKey: ks.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{ks.GetKey(o.DataSampleRef)},
			},
		},
		Inputs:  GetNewTaskInputs(ks, o.Inputs),
		Outputs: o.Outputs,
	}
}

func DefaultTrainTaskOptions() *TrainTaskOptions {
	return &TrainTaskOptions{
		KeyRef:         DefaultTrainTaskRef,
		AlgoRef:        DefaultSimpleAlgoRef,
		ParentsRef:     []string{},
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

func (o *TrainTaskOptions) WithParentsRef(p ...string) *TrainTaskOptions {
	o.ParentsRef = p
	return o
}

func (o *TrainTaskOptions) WithAlgoRef(ref string) *TrainTaskOptions {
	o.AlgoRef = ref
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

	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = ks.GetKey(ref)
	}

	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: ks.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{ks.GetKey(o.DataSampleRef)},
			},
		},
		Inputs:  GetNewTaskInputs(ks, o.Inputs),
		Outputs: o.Outputs,
	}
}

func DefaultPredictTaskOptions() *PredictTaskOptions {
	return &PredictTaskOptions{
		KeyRef:         DefaultPredictTaskRef,
		AlgoRef:        DefaultPredictAlgoRef,
		ParentsRef:     []string{},
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

func (o *PredictTaskOptions) WithParentsRef(p ...string) *PredictTaskOptions {
	o.ParentsRef = p
	return o
}

func (o *PredictTaskOptions) WithAlgoRef(ref string) *PredictTaskOptions {
	o.AlgoRef = ref
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
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = ks.GetKey(ref)
	}

	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_PREDICT,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Predict{
			Predict: &asset.NewPredictTaskData{
				DataManagerKey: ks.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{ks.GetKey(o.DataSampleRef)},
			},
		},
		Inputs:  GetNewTaskInputs(ks, o.Inputs),
		Outputs: o.Outputs,
	}
}

func DefaultCompositeTaskOptions() *CompositeTaskOptions {
	return &CompositeTaskOptions{
		KeyRef:         DefaultCompositeTaskRef,
		AlgoRef:        DefaultCompositeAlgoRef,
		ParentsRef:     []string{},
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

func (o *CompositeTaskOptions) WithParentsRef(p ...string) *CompositeTaskOptions {
	o.ParentsRef = p
	return o
}

func (o *CompositeTaskOptions) WithAlgoRef(ref string) *CompositeTaskOptions {
	o.AlgoRef = ref
	return o
}

func (o *CompositeTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = ks.GetKey(ref)
	}
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_COMPOSITE,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Composite{
			Composite: &asset.NewCompositeTrainTaskData{
				DataManagerKey: ks.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{ks.GetKey(o.DataSampleRef)},
			},
		},
		Inputs:  GetNewTaskInputs(ks, o.Inputs),
		Outputs: o.Outputs,
	}
}

func DefaultAggregateTaskOptions() *AggregateTaskOptions {
	return &AggregateTaskOptions{
		KeyRef:     DefaultAggregateTaskRef,
		AlgoRef:    DefaultAggregateAlgoRef,
		ParentsRef: []string{},
		PlanRef:    DefaultPlanRef,
		Worker:     "MyOrg1MSP",
		Inputs:     []*TaskInputOptions{},
		Outputs: map[string]*asset.NewComputeTaskOutput{
			"model": {Permissions: &asset.NewPermissions{Public: true}},
		},
	}
}

func (o *AggregateTaskOptions) WithKeyRef(ref string) *AggregateTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *AggregateTaskOptions) WithParentsRef(p ...string) *AggregateTaskOptions {
	o.ParentsRef = p
	return o
}

func (o *AggregateTaskOptions) WithWorker(w string) *AggregateTaskOptions {
	o.Worker = w
	return o
}

func (o *AggregateTaskOptions) WithAlgoRef(ref string) *AggregateTaskOptions {
	o.AlgoRef = ref
	return o
}

func (o *AggregateTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = ks.GetKey(ref)
	}
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_AGGREGATE,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Worker:         o.Worker,
		Data: &asset.NewComputeTask_Aggregate{
			Aggregate: &asset.NewAggregateTrainTaskData{},
		},
		Inputs:  GetNewTaskInputs(ks, o.Inputs),
		Outputs: o.Outputs,
	}
}

func DefaultSimpleAlgoOptions() *AlgoOptions {
	return &AlgoOptions{
		KeyRef:   DefaultSimpleAlgoRef,
		Category: asset.AlgoCategory_ALGO_SIMPLE,
		Inputs: map[string]*asset.AlgoInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"model":       {Kind: asset.AssetKind_ASSET_MODEL, Multiple: true, Optional: true},
		},
		Outputs: map[string]*asset.AlgoOutput{
			"model": {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
}

func DefaultCompositeAlgoOptions() *AlgoOptions {
	return &AlgoOptions{
		KeyRef:   DefaultCompositeAlgoRef,
		Category: asset.AlgoCategory_ALGO_COMPOSITE,
		Inputs: map[string]*asset.AlgoInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"shared":      {Kind: asset.AssetKind_ASSET_MODEL, Optional: true},
			"local":       {Kind: asset.AssetKind_ASSET_MODEL, Optional: true},
		},
		Outputs: map[string]*asset.AlgoOutput{
			"shared": {Kind: asset.AssetKind_ASSET_MODEL},
			"local":  {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
}

func DefaultAggregateAlgoOptions() *AlgoOptions {
	return &AlgoOptions{
		KeyRef:   DefaultAggregateAlgoRef,
		Category: asset.AlgoCategory_ALGO_AGGREGATE,
		Inputs: map[string]*asset.AlgoInput{
			"model": {Kind: asset.AssetKind_ASSET_MODEL, Multiple: true, Optional: true},
		},
		Outputs: map[string]*asset.AlgoOutput{
			"model": {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
}

func DefaultPredictAlgoOptions() *AlgoOptions {
	return &AlgoOptions{
		KeyRef:   DefaultPredictAlgoRef,
		Category: asset.AlgoCategory_ALGO_PREDICT,
		Inputs: map[string]*asset.AlgoInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"model":       {Kind: asset.AssetKind_ASSET_MODEL},
		},
		Outputs: map[string]*asset.AlgoOutput{
			"predictions": {Kind: asset.AssetKind_ASSET_MODEL},
		},
	}
}

func DefaultMetricAlgoOptions() *AlgoOptions {
	return &AlgoOptions{
		KeyRef:   DefaultMetricAlgoRef,
		Category: asset.AlgoCategory_ALGO_METRIC,
		Inputs: map[string]*asset.AlgoInput{
			"opener":      {Kind: asset.AssetKind_ASSET_DATA_MANAGER},
			"datasamples": {Kind: asset.AssetKind_ASSET_DATA_SAMPLE, Multiple: true},
			"predictions": {Kind: asset.AssetKind_ASSET_MODEL},
		},
		Outputs: map[string]*asset.AlgoOutput{
			"performance": {Kind: asset.AssetKind_ASSET_PERFORMANCE},
		},
	}
}

func (o *AlgoOptions) WithOutput(identifier string, kind asset.AssetKind, multiple bool) *AlgoOptions {
	o.Outputs[identifier] = &asset.AlgoOutput{
		Kind:     kind,
		Multiple: multiple,
	}
	return o
}

// SetInputs will override existing inputs with provided argument
func (o *AlgoOptions) SetInputs(inputs map[string]*asset.AlgoInput) *AlgoOptions {
	o.Inputs = inputs
	return o
}

// SetOutputs will override existing outputs with provided argument
func (o *AlgoOptions) SetOutputs(outputs map[string]*asset.AlgoOutput) *AlgoOptions {
	o.Outputs = outputs
	return o
}

func (o *AlgoOptions) WithKeyRef(ref string) *AlgoOptions {
	o.KeyRef = ref
	return o
}

func (o *AlgoOptions) WithCategory(category asset.AlgoCategory) *AlgoOptions {
	o.Category = category
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
		Category:   asset.ModelCategory_MODEL_SIMPLE,
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

func (o *ModelOptions) WithCategory(category asset.ModelCategory) *ModelOptions {
	o.Category = category
	return o
}

func DefaultPerformanceOptions() *PerformanceOptions {
	return &PerformanceOptions{
		ComputeTaskKeyRef: DefaultTrainTaskRef,
		ComputeTaskOutput: "performance",
		MetricKeyRef:      DefaultMetricAlgoRef,
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

func (o *PerformanceOptions) WithMetricRef(ref string) *PerformanceOptions {
	o.MetricKeyRef = ref
	return o
}

func DefaultDataSampleOptions() *DataSampleOptions {
	return &DataSampleOptions{
		KeyRef:   "ds",
		TestOnly: false,
	}
}

func (o *DataSampleOptions) WithKeyRef(ref string) *DataSampleOptions {
	o.KeyRef = ref
	return o
}

func (o *DataSampleOptions) WithTestOnly(flag bool) *DataSampleOptions {
	o.TestOnly = flag
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
