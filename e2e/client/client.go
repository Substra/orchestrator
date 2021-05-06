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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func NewTestClient(conn *grpc.ClientConn, mspid, channel, chaincode string) (*TestClient, error) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "mspid", mspid, "channel", channel, "chaincode", chaincode)

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

// EnsureNode attempts to register the node but won't fail on existing node
func (c *TestClient) EnsureNode() {
	_, err := c.nodeService.RegisterNode(c.ctx, &asset.NodeRegistrationParam{})
	if status.Code(err) == codes.AlreadyExists {
		log.Debug("node already exists")
		// expected error
		return
	}
	if err != nil {
		log.WithError(err).Fatal("RegisterNode failed")
	}
}

func (c *TestClient) RegisterAlgo(o *AlgoOptions) {
	newAlgo := &asset.NewAlgo{
		Key:      c.GetKey(o.KeyRef),
		Name:     "Algo test",
		Category: o.Category,
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc",
		},
		Algorithm: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/algo",
		},
		NewPermissions: &asset.NewPermissions{Public: true},
	}
	log.WithField("algo", newAlgo).Debug("registering algo")
	_, err := c.algoService.RegisterAlgo(c.ctx, newAlgo)
	if err != nil {
		log.WithError(err).Fatal("RegisterAlgo failed")
	}

}

func (c *TestClient) RegisterDataManager() {
	newDm := &asset.NewDataManager{
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
	}
	log.WithField("datamanager", newDm).Debug("registering datamanager")
	_, err := c.dataManagerService.RegisterDataManager(c.ctx, newDm)
	if err != nil {
		log.WithError(err).Fatal("RegisterDataManager failed")
	}

}

func (c *TestClient) RegisterDataSample() {
	newDs := &asset.NewDataSample{
		Keys:            []string{c.GetKey("ds")},
		DataManagerKeys: []string{c.GetKey("dm")},
		TestOnly:        false,
	}
	log.WithField("datasample", newDs).Debug("registering datasample")
	_, err := c.dataSampleService.RegisterDataSample(c.ctx, newDs)
	if err != nil {
		log.WithError(err).Fatal("RegisterDataSample failed")
	}

}

func (c *TestClient) RegisterTrainTask(o *TrainTaskOptions) {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = c.GetKey(ref)
	}
	newTask := &asset.NewComputeTask{
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
	}
	log.WithField("task", newTask).Debug("registering train task")
	_, err := c.computeTaskService.RegisterTask(c.ctx, newTask)
	if err != nil {
		log.WithError(err).Fatal("RegisterComputeTask failed")
	}

}

func (c *TestClient) RegisterCompositeTask(o *CompositeTaskOptions) {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = c.GetKey(ref)
	}
	newTask := &asset.NewComputeTask{
		Key:            c.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_COMPOSITE,
		AlgoKey:        c.GetKey(o.AlgoRef),
		ParentTaskKeys: parentKeys,
		ComputePlanKey: c.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Composite{
			Composite: &asset.NewCompositeTrainTaskData{
				DataManagerKey:   c.GetKey(o.DataManagerRef),
				DataSampleKeys:   []string{c.GetKey(o.DataSampleRef)},
				TrunkPermissions: &asset.NewPermissions{Public: true},
			},
		},
	}

	log.WithField("task", newTask).Debug("registering composite task")
	_, err := c.computeTaskService.RegisterTask(c.ctx, newTask)
	if err != nil {
		log.WithError(err).Fatal("RegisterCompositeTask failed")
	}

}

func (c *TestClient) RegisterAggregateTask(o *AggregateTaskOptions) {
	parentKeys := make([]string, len(o.ParentsRef))
	for i, ref := range o.ParentsRef {
		parentKeys[i] = c.GetKey(ref)
	}
	newTask := &asset.NewComputeTask{
		Key:            c.GetKey(o.KeyRef),
		Category:       asset.ComputeTaskCategory_TASK_AGGREGATE,
		AlgoKey:        c.GetKey(o.AlgoRef),
		ParentTaskKeys: parentKeys,
		ComputePlanKey: c.GetKey(o.PlanRef),
		Data: &asset.NewComputeTask_Aggregate{
			Aggregate: &asset.NewAggregateTrainTaskData{
				Worker: o.Worker,
			},
		},
	}

	log.WithField("task", newTask).Debug("registering aggregate task")
	_, err := c.computeTaskService.RegisterTask(c.ctx, newTask)
	if err != nil {
		log.WithError(err).Fatal("RegisterAggregateTask failed")
	}

}

func (c *TestClient) StartTask(keyRef string) {
	c.applyTaskAction(keyRef, asset.ComputeTaskAction_TASK_ACTION_DOING)
}

func (c *TestClient) DoneTask(keyRef string) {
	c.applyTaskAction(keyRef, asset.ComputeTaskAction_TASK_ACTION_DONE)
}

func (c *TestClient) CancelTask(keyRef string) {
	c.applyTaskAction(keyRef, asset.ComputeTaskAction_TASK_ACTION_CANCELED)
}

func (c *TestClient) FailTask(keyRef string) {
	c.applyTaskAction(keyRef, asset.ComputeTaskAction_TASK_ACTION_FAILED)
}

func (c *TestClient) applyTaskAction(keyRef string, action asset.ComputeTaskAction) {
	taskKey := c.GetKey(keyRef)
	log.WithField("taskKey", taskKey).WithField("action", action).Debug("applying task action")
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: taskKey,
		Action:         action,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to mark task as FAILED")
	}
}

func (c *TestClient) RegisterModel(o *ModelOptions) {
	newModel := &asset.NewModel{
		ComputeTaskKey: c.GetKey(o.TaskRef),
		Key:            c.GetKey(o.KeyRef),
		Category:       o.Category,
		Address: &asset.Addressable{
			Checksum:       "5e12e1a2687d81b268558217856547f8a4519f9688933351386a7f902cf1ce5d",
			StorageAddress: "http://somewhere.online/model",
		},
	}
	log.WithField("model", newModel).Debug("registering model")
	_, err := c.modelService.RegisterModel(c.ctx, newModel)
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
	modelKey := c.GetKey(modelRef)
	log.WithField("modelKey", modelKey).Debug("disabling model")
	_, err := c.modelService.DisableModel(c.ctx, &asset.DisableModelParam{ModelKey: modelKey})
	if err != nil {
		log.WithError(err).Fatal("DisableModel failed")
	}
}

func (c *TestClient) RegisterComputePlan(o *ComputePlanOptions) {
	newCp := &asset.NewComputePlan{
		Key:                      c.GetKey(o.KeyRef),
		DeleteIntermediaryModels: o.DeleteIntermediaryModels,
	}
	log.WithField("plan", newCp).Debug("registering compute plan")
	_, err := c.computePlanService.RegisterPlan(c.ctx, newCp)
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
