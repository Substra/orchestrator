//go:build e2e
// +build e2e

package client

func WithInput[T GenericTaskOptions](o T, identifier string, taskOutputRef *TaskOutputRef) T {
	return WithTaskInputOptions(o, &TaskInputOptions{
		Identifier: identifier,
		TaskOutput: taskOutputRef,
	})
}

func WithInputAsset[T GenericTaskOptions](o T, identifier string, assetRef string) T {
	return WithTaskInputOptions(o, &TaskInputOptions{
		Identifier: identifier,
		AssetRef:   assetRef,
	})
}

func WithTaskInputOptions[T GenericTaskOptions](o T, input *TaskInputOptions) T {
	inputs := append(o.GetInputs(), input)
	o.SetInputs(inputs)
	return o
}
