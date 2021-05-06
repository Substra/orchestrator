// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import "github.com/owkin/orchestrator/lib/asset"

type ComputePlanOptions struct {
	KeyRef                   string
	DeleteIntermediaryModels bool
}

type AlgoOptions struct {
	KeyRef   string
	Category asset.AlgoCategory
}

type TrainTaskOptions struct {
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
}

type AggregateTaskOptions struct {
	KeyRef     string
	AlgoRef    string
	ParentsRef []string
	PlanRef    string
	Worker     string
}

type ModelOptions struct {
	KeyRef   string
	TaskRef  string
	Category asset.ModelCategory
}

func DefaultTrainTaskOptions() *TrainTaskOptions {
	return &TrainTaskOptions{
		KeyRef:         DefaultTaskRef,
		AlgoRef:        "algo",
		ParentsRef:     []string{},
		PlanRef:        "cp",
		DataManagerRef: "dm",
		DataSampleRef:  "ds",
	}
}

func (o *TrainTaskOptions) WithKeyRef(ref string) *TrainTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *TrainTaskOptions) WithParentsRef(p []string) *TrainTaskOptions {
	o.ParentsRef = p
	return o
}

func (o *TrainTaskOptions) WithAlgoRef(ref string) *TrainTaskOptions {
	o.AlgoRef = ref
	return o
}

func DefaultCompositeTaskOptions() *CompositeTaskOptions {
	return &CompositeTaskOptions{
		KeyRef:         DefaultTaskRef,
		AlgoRef:        "algo",
		ParentsRef:     []string{},
		PlanRef:        "cp",
		DataManagerRef: "dm",
		DataSampleRef:  "ds",
	}
}

func (o *CompositeTaskOptions) WithKeyRef(ref string) *CompositeTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *CompositeTaskOptions) WithParentsRef(p []string) *CompositeTaskOptions {
	o.ParentsRef = p
	return o
}

func (o *CompositeTaskOptions) WithAlgoRef(ref string) *CompositeTaskOptions {
	o.AlgoRef = ref
	return o
}

func DefaultAggregateTaskOptions() *AggregateTaskOptions {
	return &AggregateTaskOptions{
		KeyRef:     DefaultTaskRef,
		AlgoRef:    "algo",
		ParentsRef: []string{},
		PlanRef:    "cp",
		Worker:     "MyOrg1MSP",
	}
}

func (o *AggregateTaskOptions) WithKeyRef(ref string) *AggregateTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *AggregateTaskOptions) WithParentsRef(p []string) *AggregateTaskOptions {
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

func DefaultAlgoOptions() *AlgoOptions {
	return &AlgoOptions{
		KeyRef:   "algo",
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
		KeyRef:                   "cp",
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
		KeyRef:   "model",
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
