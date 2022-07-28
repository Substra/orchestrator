//go:build e2e
// +build e2e

// This file contains boilerplate code. It can be deleted once we have only one ComputeTaskOptions class instead of
// {Train,Test,...}TaskOptions.

package client

type GenericTaskOptions interface {
	GetInputs() []*TaskInputOptions
	SetInputs(inputs []*TaskInputOptions)
}

// Train

func (o *TrainTaskOptions) GetInputs() []*TaskInputOptions {
	return o.Inputs
}

func (o *TrainTaskOptions) SetInputs(inputs []*TaskInputOptions) {
	o.Inputs = inputs
}

func (o *TrainTaskOptions) WithInput(identifier string, taskOutputRef *TaskOutputRef) *TrainTaskOptions {
	return WithInput(o, identifier, taskOutputRef)
}

func (o *TrainTaskOptions) WithInputAsset(identifier string, assetRef string) *TrainTaskOptions {
	return WithInputAsset(o, identifier, assetRef)
}

// Predict

func (o *PredictTaskOptions) GetInputs() []*TaskInputOptions {
	return o.Inputs
}

func (o *PredictTaskOptions) SetInputs(inputs []*TaskInputOptions) {
	o.Inputs = inputs
}

func (o *PredictTaskOptions) WithInput(identifier string, taskOutputRef *TaskOutputRef) *PredictTaskOptions {
	return WithInput(o, identifier, taskOutputRef)
}

func (o *PredictTaskOptions) WithInputAsset(identifier string, assetRef string) *PredictTaskOptions {
	return WithInputAsset(o, identifier, assetRef)
}

// Test

func (o *TestTaskOptions) GetInputs() []*TaskInputOptions {
	return o.Inputs
}

func (o *TestTaskOptions) SetInputs(inputs []*TaskInputOptions) {
	o.Inputs = inputs
}

func (o *TestTaskOptions) WithInput(identifier string, taskOutputRef *TaskOutputRef) *TestTaskOptions {
	return WithInput(o, identifier, taskOutputRef)
}

func (o *TestTaskOptions) WithInputAsset(identifier string, assetRef string) *TestTaskOptions {
	return WithInputAsset(o, identifier, assetRef)
}

// Composite

func (o *CompositeTaskOptions) GetInputs() []*TaskInputOptions {
	return o.Inputs
}

func (o *CompositeTaskOptions) SetInputs(inputs []*TaskInputOptions) {
	o.Inputs = inputs
}

func (o *CompositeTaskOptions) WithInput(identifier string, taskOutputRef *TaskOutputRef) *CompositeTaskOptions {
	return WithInput(o, identifier, taskOutputRef)
}

func (o *CompositeTaskOptions) WithInputAsset(identifier string, assetRef string) *CompositeTaskOptions {
	return WithInputAsset(o, identifier, assetRef)
}

// Aggregate

func (o *AggregateTaskOptions) GetInputs() []*TaskInputOptions {
	return o.Inputs
}

func (o *AggregateTaskOptions) SetInputs(inputs []*TaskInputOptions) {
	o.Inputs = inputs
}

func (o *AggregateTaskOptions) WithInput(identifier string, taskOutputRef *TaskOutputRef) *AggregateTaskOptions {
	return WithInput(o, identifier, taskOutputRef)
}

func (o *AggregateTaskOptions) WithInputAsset(identifier string, assetRef string) *AggregateTaskOptions {
	return WithInputAsset(o, identifier, assetRef)
}
