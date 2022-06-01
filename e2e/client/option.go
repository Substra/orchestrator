//go:build e2e
// +build e2e

package client

import (
	"github.com/go-playground/log/v7"

	"github.com/owkin/orchestrator/lib/asset"
)

type ComputePlanOptions struct {
	KeyRef                   string
	DeleteIntermediaryModels bool
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

type TaskInputOptions struct {
	Identifier                 string
	AssetRef                   string
	ParentTaskRef              string
	ParentTaskOutputIdentifier string
}

type TestTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	MetricsRef     []string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
}

type TrainTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
}

type PredictTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
}

type CompositeTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
	Inputs         []*TaskInputOptions
}

type AggregateTaskOptions struct {
	KeyRef     string
	AlgoRef    string
	ParentsRef []string
	PlanRef    string
	Worker     string
	Inputs     []*TaskInputOptions
}

type ModelOptions struct {
	KeyRef   string
	TaskRef  string
	Category asset.ModelCategory
}

type PerformanceOptions struct {
	ComputeTaskKeyRef string
	MetricKeyRef      string
	PerformanceValue  float32
}

type DataManagerOptions struct {
	LogsPermission *asset.NewPermissions
}

func DefaultTestTaskOptions() *TestTaskOptions {
	return &TestTaskOptions{
		KeyRef:         DefaultTaskRef,
		AlgoRef:        DefaultAlgoRef,
		ParentsRef:     []string{},
		PlanRef:        DefaultPlanRef,
		MetricsRef:     []string{DefaultMetricRef},
		DataManagerRef: DefaultDataManagerRef,
		DataSampleRef:  DefaultDataSampleRef,
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

func (o *TestTaskOptions) WithMetricsRef(ref ...string) *TestTaskOptions {
	o.MetricsRef = ref
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
	metricKeys := make([]string, len(o.MetricsRef))
	for i, ref := range o.MetricsRef {
		metricKeys[i] = ks.GetKey(ref)
	}
	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_TEST,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ParentTaskKeys: parentKeys,
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Test{
			Test: &asset.NewTestTaskData{
				DataManagerKey: ks.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{ks.GetKey(o.DataSampleRef)},
				MetricKeys:     metricKeys,
			},
		},
	}
}

func DefaultTrainTaskOptions() *TrainTaskOptions {
	return &TrainTaskOptions{
		KeyRef:         DefaultTaskRef,
		AlgoRef:        DefaultAlgoRef,
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

func (o *TrainTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {

	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = ks.GetKey(ref)
	}

	inputs := make([]*asset.ComputeTaskInput, len(o.Inputs))
	for i, in := range o.Inputs {

		inputs[i] = &asset.ComputeTaskInput{
			Identifier: in.Identifier,
		}

		if (in.AssetRef != "") && (in.ParentTaskRef != "" || in.ParentTaskOutputIdentifier != "") {
			log.Fatal("Cannot have AssetRef and (ParentTaskRef or OutputIdentifier) at the same time.")
		}

		if in.AssetRef != "" {
			inputs[i].Ref = &asset.ComputeTaskInput_AssetKey{
				AssetKey: ks.GetKey(in.AssetRef),
			}
		} else if in.ParentTaskRef != "" {
			inputs[i].Ref = &asset.ComputeTaskInput_ParentTaskOutput{
				ParentTaskOutput: &asset.ParentTaskOutputRef{
					ParentTaskKey:    ks.GetKey(in.ParentTaskRef),
					OutputIdentifier: in.ParentTaskOutputIdentifier,
				},
			}
		}
	}

	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ParentTaskKeys: parentKeys,
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: ks.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{ks.GetKey(o.DataSampleRef)},
			},
		},
		Inputs: inputs,
	}
}

func DefaultPredictTaskOptions() *PredictTaskOptions {
	return &PredictTaskOptions{
		KeyRef:         DefaultTaskRef,
		AlgoRef:        DefaultAlgoRef,
		ParentsRef:     []string{},
		PlanRef:        DefaultPlanRef,
		DataManagerRef: DefaultDataManagerRef,
		DataSampleRef:  DefaultDataSampleRef,
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

func (o *PredictTaskOptions) GetNewTask(ks *KeyStore) *asset.NewComputeTask {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = ks.GetKey(ref)
	}

	return &asset.NewComputeTask{
		Key:            ks.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_PREDICT,
		AlgoKey:        ks.GetKey(o.AlgoRef),
		ParentTaskKeys: parentKeys,
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Predict{
			Predict: &asset.NewPredictTaskData{
				DataManagerKey: ks.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{ks.GetKey(o.DataSampleRef)},
			},
		},
	}
}

func DefaultCompositeTaskOptions() *CompositeTaskOptions {
	return &CompositeTaskOptions{
		KeyRef:         DefaultTaskRef,
		AlgoRef:        DefaultAlgoRef,
		ParentsRef:     []string{},
		PlanRef:        DefaultPlanRef,
		DataManagerRef: DefaultDataManagerRef,
		DataSampleRef:  DefaultDataSampleRef,
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
		ParentTaskKeys: parentKeys,
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Composite{
			Composite: &asset.NewCompositeTrainTaskData{
				DataManagerKey:   ks.GetKey(o.DataManagerRef),
				DataSampleKeys:   []string{ks.GetKey(o.DataSampleRef)},
				TrunkPermissions: &asset.NewPermissions{Public: true},
			},
		},
	}
}

func DefaultAggregateTaskOptions() *AggregateTaskOptions {
	return &AggregateTaskOptions{
		KeyRef:     DefaultTaskRef,
		AlgoRef:    DefaultAlgoRef,
		ParentsRef: []string{},
		PlanRef:    DefaultPlanRef,
		Worker:     "MyOrg1MSP",
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

func (o *TrainTaskOptions) WithParentTaskInputRef(identifier string, parentTaskRef string, parentTaskOutputIdentifier string) *TrainTaskOptions {
	o.Inputs = append(o.Inputs, &TaskInputOptions{
		Identifier:                 identifier,
		ParentTaskRef:              parentTaskRef,
		ParentTaskOutputIdentifier: parentTaskOutputIdentifier,
	})
	return o
}

func (o *TrainTaskOptions) WithAssetInputRef(identifier string, assetRef string) *TrainTaskOptions {
	o.Inputs = append(o.Inputs, &TaskInputOptions{
		Identifier: identifier,
		AssetRef:   assetRef,
	})
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
		ParentTaskKeys: parentKeys,
		ComputePlanKey: ks.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Aggregate{
			Aggregate: &asset.NewAggregateTrainTaskData{
				Worker: o.Worker,
			},
		},
	}
}

func DefaultAlgoOptions() *AlgoOptions {
	return &AlgoOptions{
		KeyRef:   DefaultAlgoRef,
		Category: asset.AlgoCategory_ALGO_SIMPLE,
	}
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
		KeyRef:                   DefaultPlanRef,
		DeleteIntermediaryModels: false,
	}
}

func (o *ComputePlanOptions) WithKeyRef(ref string) *ComputePlanOptions {
	o.KeyRef = ref
	return o
}

func (o *ComputePlanOptions) WithDeleteIntermediaryModels(flag bool) *ComputePlanOptions {
	o.DeleteIntermediaryModels = flag
	return o
}

func DefaultModelOptions() *ModelOptions {
	return &ModelOptions{
		KeyRef:   DefaultModelRef,
		TaskRef:  DefaultTaskRef,
		Category: asset.ModelCategory_MODEL_SIMPLE,
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

func (o *ModelOptions) WithCategory(category asset.ModelCategory) *ModelOptions {
	o.Category = category
	return o
}

func DefaultPerformanceOptions() *PerformanceOptions {
	return &PerformanceOptions{
		ComputeTaskKeyRef: DefaultTaskRef,
		MetricKeyRef:      DefaultMetricRef,
		PerformanceValue:  0.5,
	}
}

func (o *PerformanceOptions) WithTaskRef(ref string) *PerformanceOptions {
	o.ComputeTaskKeyRef = ref
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
