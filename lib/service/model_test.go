package service

import (
	"errors"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetComputeTasksOutputModels(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetModelDBAL").Return(dbal)

	service := NewModelService(provider)

	returnedModels := []*asset.Model{{}, {}, {}}

	dbal.On("GetComputeTaskOutputModels", "taskUuid").Once().Return(returnedModels, nil)

	models, err := service.GetComputeTaskOutputModels("taskUuid")
	assert.NoError(t, err)

	assert.Len(t, models, 3)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetModelDBAL").Return(dbal)

	service := NewModelService(provider)

	model := &asset.Model{
		Key: "uuid",
	}

	dbal.On("GetModel", "uuid").Once().Return(model, nil)

	ret, err := service.GetModel("uuid")
	assert.NoError(t, err)
	assert.Equal(t, model, ret)

	provider.AssertExpectations(t)
	dbal.AssertExpectations(t)
}

func TestGetCheckedModel(t *testing.T) {
	model := &asset.Model{
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        false,
				AuthorizedIds: []string{"worker"},
			},
		},
	}

	dbal := new(persistence.MockDBAL)
	dbal.On("GetModel", "uuid").Return(model, nil)
	dbal.On("GetModel", "unknown uuid").Return(nil, orcerrors.NewNotFound("model", "unknown uuid"))

	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(NewPermissionService(provider))

	service := NewModelService(provider)

	var actual *asset.Model
	var err error

	actual, err = service.GetCheckedModel("uuid", "worker")
	assert.NoError(t, err)
	assert.Equal(t, model, actual)

	actual, err = service.GetCheckedModel("unknown uuid", "worker")
	assert.ErrorContains(t, err, "not found")
	assert.Nil(t, actual)

	actual, err = service.GetCheckedModel("uuid", "bad worker")
	assert.ErrorContains(t, err, "not authorized")
	assert.Nil(t, actual)
}

func TestRegisterOnNonDoingTask(t *testing.T) {
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	service := NewModelService(provider)

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "test",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{},
		&asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "test",
		})
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrBadRequest, orcError.Kind)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterModelWrongPermissions(t *testing.T) {
	provider := newMockedProvider()
	service := NewModelService(provider)

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "test",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{},
		&asset.ComputeTask{
			Status: asset.ComputeTaskStatus_STATUS_DONE,
			Worker: "owner",
		}) // "test" is not "owner" of the task
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrPermissionDenied, orcError.Kind)

	provider.AssertExpectations(t)
}

func TestRegisterTrainModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
		Worker:   "test",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				},
			},
		},
		Algo: &asset.Algo{
			Outputs: map[string]*asset.AlgoOutput{
				"model": {
					Kind: asset.AssetKind_ASSET_MODEL,
				},
			},
		},
	}

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "model",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	storedModel := &asset.Model{
		Key:            model.Key,
		Category:       model.Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        model.Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
		Owner:        "test",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddModel", storedModel, "model").Once().Return(nil)

	output := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              model.ComputeTaskKey,
		ComputeTaskOutputIdentifier: model.ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    model.Key,
	}
	cts.On("addComputeTaskOutputAsset", output).Once().Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  model.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_Model{Model: storedModel},
	}
	es.On("RegisterEvents", event).Once().Return(nil)

	_, err := service.registerModel(model, "test", persistence.ComputeTaskOutputCounter{}, task)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterAggregateModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_AGGREGATE,
		Worker:   "test",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{"org1", "org2"},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				},
			},
		},
		Algo: &asset.Algo{
			Outputs: map[string]*asset.AlgoOutput{
				"model": {
					Kind: asset.AssetKind_ASSET_MODEL,
				},
			},
		},
	}

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "model",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	storedModel := &asset.Model{
		Key:            model.Key,
		Category:       model.Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        model.Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{"org1", "org2"},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
		Owner:        "test",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddModel", storedModel, "model").Once().Return(nil)

	output := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              model.ComputeTaskKey,
		ComputeTaskOutputIdentifier: model.ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    model.Key,
	}
	cts.On("addComputeTaskOutputAsset", output).Once().Return(nil)

	es.On("RegisterEvents", mock.AnythingOfType("*asset.Event")).Once().Return(nil)

	_, err := service.registerModel(model, "test", persistence.ComputeTaskOutputCounter{}, task)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterDuplicateModel(t *testing.T) {
	provider := newMockedProvider()
	service := NewModelService(provider)

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "model",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{"model": 1},
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_TRAIN,
			Worker:   "test",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"model": {
					Permissions: &asset.Permissions{
						Process: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{"org1", "org2"},
						},
						Download: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{},
						},
					},
				},
			},
			Algo: &asset.Algo{
				Outputs: map[string]*asset.AlgoOutput{
					"model": {
						Kind: asset.AssetKind_ASSET_MODEL,
					},
				},
			},
		},
	)
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrConflict, orcError.Kind)

	provider.AssertExpectations(t)
}

func TestRegisterHeadModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
		Worker:   "test",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"shared": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				}},
			"local": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				}},
		},
		Algo: &asset.Algo{
			Outputs: map[string]*asset.AlgoOutput{
				"local": {
					Kind: asset.AssetKind_ASSET_MODEL,
				},
				"shared": {
					Kind: asset.AssetKind_ASSET_MODEL,
				},
			},
		},
	}

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_HEAD,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "local",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	storedModel := &asset.Model{
		Key:            model.Key,
		Category:       model.Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        model.Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
		Owner:        "test",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddModel", storedModel, "local").Once().Return(nil)

	output := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              model.ComputeTaskKey,
		ComputeTaskOutputIdentifier: model.ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    model.Key,
	}
	cts.On("addComputeTaskOutputAsset", output).Once().Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  model.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_Model{Model: storedModel},
	}
	es.On("RegisterEvents", event).Once().Return(nil)

	_, err := service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{"shared": 1},
		task)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterWrongModelType(t *testing.T) {
	provider := newMockedProvider()
	service := NewModelService(provider)

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_HEAD, // cannot register a HEAD model on aggregate task
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "model",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{},
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_AGGREGATE,
			Worker:   "test",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"model": {
					Permissions: &asset.Permissions{
						Process: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{},
						},
						Download: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{},
						},
					}},
			},
			Algo: &asset.Algo{
				Outputs: map[string]*asset.AlgoOutput{
					"model": {
						Kind: asset.AssetKind_ASSET_MODEL,
					},
				},
			},
		})
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrBadRequest, orcError.Kind)

	provider.AssertExpectations(t)
}

func TestRegisterMultipleHeads(t *testing.T) {
	provider := newMockedProvider()
	service := NewModelService(provider)

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_HEAD,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "local",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{"local": 1},
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
			Worker:   "test",
			Outputs: map[string]*asset.ComputeTaskOutput{
				"shared": {
					Permissions: &asset.Permissions{
						Process: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{},
						},
						Download: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{},
						},
					}},
				"local": {
					Permissions: &asset.Permissions{
						Process: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{},
						},
						Download: &asset.Permission{
							Public:        true,
							AuthorizedIds: []string{},
						},
					}},
			},
			Algo: &asset.Algo{
				Outputs: map[string]*asset.AlgoOutput{
					"local": {
						Kind: asset.AssetKind_ASSET_MODEL,
					},
					"shared": {
						Kind: asset.AssetKind_ASSET_MODEL,
					},
				},
			},
		})
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrConflict, orcError.Kind)

	provider.AssertExpectations(t)
}

func TestRegisterInvalidOutput(t *testing.T) {
	provider := newMockedProvider()
	service := NewModelService(provider)

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
		Worker:   "test",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				}},
		},
		Algo: &asset.Algo{
			Outputs: map[string]*asset.AlgoOutput{
				"model": {
					Kind: asset.AssetKind_ASSET_UNKNOWN,
				},
			},
		},
	}

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_HEAD,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "invalid",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{},
		task)
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrMissingTaskOutput, orcError.Kind)

	model.ComputeTaskOutputIdentifier = "model"
	_, err = service.registerModel(
		model,
		"test",
		persistence.ComputeTaskOutputCounter{},
		task)
	assert.Error(t, err)
	orcError = new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrIncompatibleKind, orcError.Kind)

	provider.AssertExpectations(t)
}

func TestGetInputModels(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "uuid").Once().Return(
		&asset.ComputeTask{
			ParentTaskKeys: []string{"parent1", "parent2"},
		},
		nil,
	)

	model1 := &asset.Model{Key: "m1"}
	model2 := &asset.Model{Key: "m2"}

	dbal.On("GetComputeTaskOutputModels", "parent1").Once().Return([]*asset.Model{model1}, nil)
	dbal.On("GetComputeTaskOutputModels", "parent2").Once().Return([]*asset.Model{model2}, nil)

	models, err := service.GetComputeTaskInputModels("uuid")
	assert.NoError(t, err)

	assert.Equal(t, []*asset.Model{model1, model2}, models)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetCompositeInputModels(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "uuid").Once().Return(
		&asset.ComputeTask{
			Category:       asset.ComputeTaskCategory_TASK_COMPOSITE,
			ParentTaskKeys: []string{"composite"},
		},
		nil,
	)

	mc1 := &asset.Model{Key: "c1", Category: asset.ModelCategory_MODEL_HEAD}
	mc2 := &asset.Model{Key: "c2", Category: asset.ModelCategory_MODEL_SIMPLE}

	dbal.On("GetComputeTaskOutputModels", "composite").Once().Return([]*asset.Model{mc1, mc2}, nil)

	models, err := service.GetComputeTaskInputModels("uuid")
	assert.NoError(t, err)

	assert.Equal(t, []*asset.Model{mc1, mc2}, models)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetInputModelsForCompositeWithAggregateParent(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "uuid").Once().Return(
		&asset.ComputeTask{
			Category:       asset.ComputeTaskCategory_TASK_COMPOSITE,
			ParentTaskKeys: []string{"aggregate", "composite"},
		},
		nil,
	)

	mc1 := &asset.Model{Key: "c1", Category: asset.ModelCategory_MODEL_HEAD}
	mc2 := &asset.Model{Key: "c2", Category: asset.ModelCategory_MODEL_SIMPLE}
	ma := &asset.Model{Key: "a", Category: asset.ModelCategory_MODEL_SIMPLE}

	dbal.On("GetComputeTaskOutputModels", "composite").Once().Return([]*asset.Model{mc1, mc2}, nil)
	dbal.On("GetComputeTaskOutputModels", "aggregate").Once().Return([]*asset.Model{ma}, nil)

	models, err := service.GetComputeTaskInputModels("uuid")
	assert.NoError(t, err)

	assert.Equal(t, []*asset.Model{ma, mc1}, models)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetInputModelsForCompositeWithCompositeParents(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "uuid").Once().Return(
		&asset.ComputeTask{
			Category:       asset.ComputeTaskCategory_TASK_COMPOSITE,
			ParentTaskKeys: []string{"composite1", "composite2"},
		},
		nil,
	)

	mc11 := &asset.Model{Key: "c11", Category: asset.ModelCategory_MODEL_HEAD}
	mc12 := &asset.Model{Key: "c12", Category: asset.ModelCategory_MODEL_SIMPLE}
	mc21 := &asset.Model{Key: "c21", Category: asset.ModelCategory_MODEL_HEAD}
	mc22 := &asset.Model{Key: "c22", Category: asset.ModelCategory_MODEL_SIMPLE}

	dbal.On("GetComputeTaskOutputModels", "composite1").Once().Return([]*asset.Model{mc11, mc12}, nil)
	dbal.On("GetComputeTaskOutputModels", "composite2").Once().Return([]*asset.Model{mc21, mc22}, nil)

	models, err := service.GetComputeTaskInputModels("uuid")
	assert.NoError(t, err)

	assert.Equal(t, []*asset.Model{mc11, mc22}, models)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetAggregateChildInputModels(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "uuid").Once().Return(
		&asset.ComputeTask{
			Category:       asset.ComputeTaskCategory_TASK_COMPOSITE,
			ParentTaskKeys: []string{"composite", "aggregate"},
		},
		nil,
	)

	mc1 := &asset.Model{Key: "c1", Category: asset.ModelCategory_MODEL_HEAD}
	mc2 := &asset.Model{Key: "c2", Category: asset.ModelCategory_MODEL_SIMPLE}
	ma := &asset.Model{Key: "aggregate", Category: asset.ModelCategory_MODEL_SIMPLE}

	dbal.On("GetComputeTaskOutputModels", "composite").Once().Return([]*asset.Model{mc1, mc2}, nil)
	dbal.On("GetComputeTaskOutputModels", "aggregate").Once().Return([]*asset.Model{ma}, nil)

	models, err := service.GetComputeTaskInputModels("uuid")
	assert.NoError(t, err)

	assert.Equal(t, []*asset.Model{mc1, ma}, models)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetAggregateChildInoutModelsWithSimpleFirst(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "uuid").Once().Return(
		&asset.ComputeTask{
			Category:       asset.ComputeTaskCategory_TASK_COMPOSITE,
			ParentTaskKeys: []string{"composite", "aggregate"},
		},
		nil,
	)

	mc1 := &asset.Model{Key: "c1", Category: asset.ModelCategory_MODEL_HEAD}
	mc2 := &asset.Model{Key: "c2", Category: asset.ModelCategory_MODEL_SIMPLE}
	ma := &asset.Model{Key: "aggregate", Category: asset.ModelCategory_MODEL_SIMPLE}

	dbal.On("GetComputeTaskOutputModels", "composite").Once().Return([]*asset.Model{mc2, mc1}, nil)
	dbal.On("GetComputeTaskOutputModels", "aggregate").Once().Return([]*asset.Model{ma}, nil)

	models, err := service.GetComputeTaskInputModels("uuid")
	assert.NoError(t, err)

	assert.Equal(t, []*asset.Model{mc1, ma}, models)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestCanDisableModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("canDisableModels", "taskKey", "requester").Once().Return(true, nil)

	dbal.On("GetModel", "modelUuid").Once().Return(&asset.Model{
		Key:            "modelUuid",
		ComputeTaskKey: "taskKey",
	}, nil)

	can, err := service.CanDisableModel("modelUuid", "requester")
	assert.NoError(t, err)
	assert.True(t, can)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestDisableModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	es := new(MockEventAPI)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetEventService").Return(es)
	service := NewModelService(provider)

	cts.On("canDisableModels", "taskKey", "requester").Return(true, nil)

	model := &asset.Model{
		Key:            "modelUuid",
		ComputeTaskKey: "taskKey",
		Address:        &asset.Addressable{Checksum: "sha", StorageAddress: "http://there"},
	}
	dbal.On("GetModel", "modelUuid").Return(model, nil)

	updatedModel := &asset.Model{Key: "modelUuid", ComputeTaskKey: "taskKey"}
	dbal.On("UpdateModel", updatedModel).Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  "modelUuid",
		EventKind: asset.EventKind_EVENT_ASSET_DISABLED,
		Asset:     &asset.Event_Model{Model: updatedModel},
	}
	es.On("RegisterEvents", event).Once().Return(nil)

	err := service.DisableModel("modelUuid", "requester")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestQueryModels(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetModelDBAL").Return(dbal)
	service := NewModelService(provider)

	model1 := asset.Model{
		Key:      "model1",
		Category: asset.ModelCategory_MODEL_SIMPLE,
	}
	model2 := asset.Model{
		Key:      "model2",
		Category: asset.ModelCategory_MODEL_SIMPLE,
	}

	pagination := common.NewPagination("", 12)

	dbal.On("QueryModels", asset.ModelCategory_MODEL_SIMPLE, pagination).Return([]*asset.Model{&model1, &model2}, "nextPage", nil).Once()

	r, token, err := service.QueryModels(asset.ModelCategory_MODEL_SIMPLE, pagination)
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, model1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestAreAllOutputsRegistered(t *testing.T) {
	cases := map[string]struct {
		task    *asset.ComputeTask
		models  []*asset.Model
		outcome bool
	}{
		"train without model": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_TRAIN},
			models:  []*asset.Model{},
			outcome: false,
		},
		"unhandled task category": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_TEST},
			models:  []*asset.Model{},
			outcome: false,
		},
		"train with model": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_TRAIN},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_SIMPLE}},
			outcome: true,
		},
		"aggregate with model": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_AGGREGATE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_SIMPLE}},
			outcome: true,
		},
		"composite with head": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_COMPOSITE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_HEAD}},
			outcome: false,
		},
		"composite with simple": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_COMPOSITE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_SIMPLE}},
			outcome: false,
		},
		"composite with head & simple": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_COMPOSITE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_SIMPLE}, {Category: asset.ModelCategory_MODEL_HEAD}},
			outcome: true,
		},
	}

	provider := newMockedProvider()
	service := NewModelService(provider)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, service.AreAllOutputsRegistered(tc.task, tc.models))
		})
	}
}

func TestRegisterModelsTrainTask(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	service := NewModelService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
		Worker:   "test",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"model": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				},
			},
		},
		Algo: &asset.Algo{
			Outputs: map[string]*asset.AlgoOutput{
				"model": {
					Kind: asset.AssetKind_ASSET_MODEL,
				},
			},
		},
	}

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(task, nil)
	cts.On("getTaskOutputCounter", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(persistence.ComputeTaskOutputCounter{}, nil)

	models := []*asset.NewModel{
		{
			Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
			Category:                    asset.ModelCategory_MODEL_SIMPLE,
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "model",
			Address: &asset.Addressable{
				StorageAddress: "https://somewhere",
				Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
			},
		},
	}

	storedModel := &asset.Model{
		Key:            models[0].Key,
		Category:       models[0].Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        models[0].Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
		Owner:        "test",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddModel", storedModel, "model").Once().Return(nil)

	output := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              models[0].ComputeTaskKey,
		ComputeTaskOutputIdentifier: models[0].ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    models[0].Key,
	}
	cts.On("addComputeTaskOutputAsset", output).Once().Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  models[0].Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_Model{Model: storedModel},
	}
	es.On("RegisterEvents", event).Once().Return(nil)

	_, err := service.RegisterModels(models, "test")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterHeadAndTrunkModel(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	service := NewModelService(provider)

	ts.On("GetTransactionTime").Times(2).Return(time.Unix(1337, 0))

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
		Worker:   "test",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"shared": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				}},
			"local": {
				Permissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				}},
		},
		Algo: &asset.Algo{
			Outputs: map[string]*asset.AlgoOutput{
				"local": {
					Kind: asset.AssetKind_ASSET_MODEL,
				},
				"shared": {
					Kind: asset.AssetKind_ASSET_MODEL,
				},
			},
		},
	}
	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Times(2).Return(task, nil)

	models := []*asset.NewModel{
		{
			Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
			Category:                    asset.ModelCategory_MODEL_HEAD,
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "local",
			Address: &asset.Addressable{
				StorageAddress: "https://somewhere",
				Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
			},
		},
		{
			Key:                         "7d2c6aa1-18b9-4ffd-a6e3-dfdc740d64dd",
			Category:                    asset.ModelCategory_MODEL_SIMPLE,
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "shared",
			Address: &asset.Addressable{
				StorageAddress: "https://somewhere",
				Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
			},
		},
	}

	storedHead := &asset.Model{
		Key:            models[0].Key,
		Category:       models[0].Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        models[0].Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
		Owner:        "test",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddModel", storedHead, "local").Once().Return(nil)

	storedSimple := &asset.Model{
		Key:            models[1].Key,
		Category:       models[1].Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        models[0].Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
		Owner:        "test",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddModel", storedSimple, "shared").Once().Return(nil)

	cts.On("getTaskOutputCounter", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(persistence.ComputeTaskOutputCounter{}, nil)

	for _, model := range models {
		output := &asset.ComputeTaskOutputAsset{
			ComputeTaskKey:              model.ComputeTaskKey,
			ComputeTaskOutputIdentifier: model.ComputeTaskOutputIdentifier,
			AssetKind:                   asset.AssetKind_ASSET_MODEL,
			AssetKey:                    model.Key,
		}
		cts.On("addComputeTaskOutputAsset", output).Once().Return(nil)
	}

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  models[0].Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_Model{Model: storedHead},
	}
	es.On("RegisterEvents", event).Once().Return(nil)

	eventSimple := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  models[1].Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_Model{Model: storedSimple},
	}
	es.On("RegisterEvents", eventSimple).Once().Return(nil)

	_, err := service.RegisterModels(models, "test")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterMissingOutput(t *testing.T) {
	provider := newMockedProvider()
	service := NewModelService(provider)

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
		Worker:   "test",
		Outputs:  map[string]*asset.ComputeTaskOutput{},
	}

	model := &asset.NewModel{
		Key:                         "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:                    asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "model",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.registerModel(model, "test", persistence.ComputeTaskOutputCounter{}, task)
	assert.ErrorContains(t, err, "has no output named \"model\"")

	provider.AssertExpectations(t)
}
