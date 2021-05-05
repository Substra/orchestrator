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

import (
	"context"

	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const DefaultTaskRef = "task"

// TestClient is a client for the tested app
type TestClient struct {
	ctx                context.Context
	keys               map[string]string
	nodeService        asset.NodeServiceClient
	objectiveService   asset.ObjectiveServiceClient
	algoService        asset.AlgoServiceClient
	dataManagerService asset.DataManagerServiceClient
	dataSampleService  asset.DataSampleServiceClient
	modelService       asset.ModelServiceClient
	computeTaskService asset.ComputeTaskServiceClient
	computePlanService asset.ComputePlanServiceClient
}

func NewTestClient(conn *grpc.ClientConn, mspid, channel string) (*TestClient, error) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "mspid", mspid, "channel", channel)

	return &TestClient{
		ctx:                ctx,
		keys:               make(map[string]string),
		nodeService:        asset.NewNodeServiceClient(conn),
		algoService:        asset.NewAlgoServiceClient(conn),
		objectiveService:   asset.NewObjectiveServiceClient(conn),
		dataManagerService: asset.NewDataManagerServiceClient(conn),
		dataSampleService:  asset.NewDataSampleServiceClient(conn),
		modelService:       asset.NewModelServiceClient(conn),
		computeTaskService: asset.NewComputeTaskServiceClient(conn),
		computePlanService: asset.NewComputePlanServiceClient(conn),
	}, nil
}

// GetKey will create a UUID or return the previously generated one.
// This is useful when building relationships between entities.
func (c *TestClient) GetKey(id string) string {
	k, ok := c.keys[id]
	if !ok {
		k = uuid.New().String()
		c.keys[id] = k
	}

	return k
}

type ComputePlanOptions struct {
	DeleteIntermediaryModels bool
}

type RegisterTrainTaskOptions struct {
	KeyRef         string
	AlgoRef        string
	ParentsRef     []string
	PlanRef        string
	DataManagerRef string
	DataSampleRef  string
}

func DefaultRegisterTrainTaskOptions() *RegisterTrainTaskOptions {
	return &RegisterTrainTaskOptions{
		KeyRef:         DefaultTaskRef,
		AlgoRef:        "algo",
		ParentsRef:     []string{},
		PlanRef:        "cp",
		DataManagerRef: "dm",
		DataSampleRef:  "ds",
	}
}

func (o *RegisterTrainTaskOptions) WithKeyRef(ref string) *RegisterTrainTaskOptions {
	o.KeyRef = ref
	return o
}

func (o *RegisterTrainTaskOptions) WithParentsRef(p []string) *RegisterTrainTaskOptions {
	o.ParentsRef = p
	return o
}

func (c *TestClient) RegisterNode() {
	_, err := c.nodeService.RegisterNode(c.ctx, &asset.NodeRegistrationParam{})
	if err != nil {
		log.WithError(err).Fatal("RegisterNode failed")
	}
}

func (c *TestClient) RegisterAlgo() {
	_, err := c.algoService.RegisterAlgo(c.ctx, &asset.NewAlgo{
		Key:      c.GetKey("algo"),
		Name:     "Algo test",
		Category: asset.AlgoCategory_ALGO_SIMPLE,
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc",
		},
		Algorithm: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/algo",
		},
		NewPermissions: &asset.NewPermissions{Public: true},
	})
	if err != nil {
		log.WithError(err).Fatal("RegisterAlgo failed")
	}

}

func (c *TestClient) RegisterDataManager() {
	_, err := c.dataManagerService.RegisterDataManager(c.ctx, &asset.NewDataManager{
		Key:            c.GetKey("dm"),
		Name:           "Test datamanager",
		NewPermissions: &asset.NewPermissions{Public: true},
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc",
		},
		Opener: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/opener",
		},
		Type: "test",
	})
	if err != nil {
		log.WithError(err).Fatal("RegisterDataManager failed")
	}

}

func (c *TestClient) RegisterDataSample() {
	_, err := c.dataSampleService.RegisterDataSample(c.ctx, &asset.NewDataSample{
		Keys:            []string{c.GetKey("ds")},
		DataManagerKeys: []string{c.GetKey("dm")},
		TestOnly:        false,
	})
	if err != nil {
		log.WithError(err).Fatal("RegisterDataSample failed")
	}

}

func (c *TestClient) RegisterTrainTask(o *RegisterTrainTaskOptions) {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = c.GetKey(ref)
	}

	_, err := c.computeTaskService.RegisterTask(c.ctx, &asset.NewComputeTask{
		Key:            c.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        c.GetKey(o.AlgoRef),
		ParentTaskKeys: parentKeys,
		ComputePlanKey: c.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: c.GetKey(o.DataManagerRef),
				DataSampleKeys: []string{c.GetKey(o.DataSampleRef)},
			},
		},
	})
	if err != nil {
		log.WithError(err).Fatal("RegisterComputeTask failed")
	}

}

func (c *TestClient) StartTrainTask(keyRef string) {
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: c.GetKey(keyRef),
		Action:         asset.ComputeTaskAction_TASK_ACTION_DOING,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to mark task as DOING")
	}
}

func (c *TestClient) DoneTrainTask(keyRef string) {
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: c.GetKey(keyRef),
		Action:         asset.ComputeTaskAction_TASK_ACTION_DONE,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to mark task as DONE")
	}
}

func (c *TestClient) CancelTrainTask(keyRef string) {
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: c.GetKey(keyRef),
		Action:         asset.ComputeTaskAction_TASK_ACTION_CANCELED,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to mark task as CANCELED")
	}
}

func (c *TestClient) FailTrainTask(keyRef string) {
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: c.GetKey(keyRef),
		Action:         asset.ComputeTaskAction_TASK_ACTION_FAILED,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to mark task as FAILED")
	}
}

func (c *TestClient) RegisterModel(taskRef, modelRef string) {
	_, err := c.modelService.RegisterModel(c.ctx, &asset.NewModel{
		ComputeTaskKey: c.GetKey(taskRef),
		Key:            c.GetKey(modelRef),
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		Address: &asset.Addressable{
			Checksum:       "5e12e1a2687d81b268558217856547f8a4519f9688933351386a7f902cf1ce5d",
			StorageAddress: "http://somewhere.online/model",
		},
	})
	if err != nil {
		log.WithError(err).Fatal("RegisterModel failed")
	}
}

func (c *TestClient) GetTaskOutputModels(taskRef string) []*asset.Model {
	resp, err := c.modelService.GetComputeTaskOutputModels(c.ctx, &asset.GetComputeTaskModelsParam{ComputeTaskKey: c.GetKey(taskRef)})
	if err != nil {
		log.WithError(err).Fatal("GetComputeTaskOutputModels failed")
	}

	return resp.Models
}

func (c *TestClient) CanDisableModel(modelRef string) bool {
	resp, err := c.modelService.CanDisableModel(c.ctx, &asset.CanDisableModelParam{ModelKey: c.GetKey(modelRef)})
	if err != nil {
		log.WithError(err).Fatal("CanDisableModel failed")
	}

	return resp.CanDisable
}

func (c *TestClient) DisableModel(modelRef string) {
	_, err := c.modelService.DisableModel(c.ctx, &asset.DisableModelParam{ModelKey: c.GetKey(modelRef)})
	if err != nil {
		log.WithError(err).Fatal("CanDisableModel failed")
	}
}

func (c *TestClient) RegisterComputePlan(opts *ComputePlanOptions) {
	newCP := &asset.NewComputePlan{
		Key: c.GetKey("cp"),
	}

	if opts != nil {
		newCP.DeleteIntermediaryModels = opts.DeleteIntermediaryModels
	}

	_, err := c.computePlanService.RegisterPlan(c.ctx, newCP)
	if err != nil {
		log.WithError(err).Fatal("RegisterPlan failed")
	}
}

func (c *TestClient) GetComputePlan(keyRef string) *asset.ComputePlan {
	plan, err := c.computePlanService.GetPlan(c.ctx, &asset.GetComputePlanParam{Key: c.GetKey(keyRef)})
	if err != nil {
		log.WithError(err).Fatalf("QueryPlans failed")
	}

	return plan
}

func (c TestClient) GetComputeTask(keyRef string) *asset.ComputeTask {
	task, err := c.computeTaskService.GetTask(c.ctx, &asset.TaskQueryParam{Key: c.GetKey(keyRef)})
	if err != nil {
		log.WithError(err).Fatalf("GetTask failed")
	}

	return task
}
