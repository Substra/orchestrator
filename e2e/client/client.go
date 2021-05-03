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
	"log"

	"github.com/google/uuid"
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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
	Plans              []*asset.ComputePlan
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

func (c *TestClient) RegisterNode() {
	_, err := c.nodeService.RegisterNode(c.ctx, &asset.NodeRegistrationParam{})
	if err != nil {
		log.Fatalf("RegisterNode failed: %v", err)
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
		log.Fatalf("RegisterAlgo failed: %v", err)
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
		log.Fatalf("RegisterDataManager failed: %v", err)
	}

}

func (c *TestClient) RegisterDataSample() {
	_, err := c.dataSampleService.RegisterDataSample(c.ctx, &asset.NewDataSample{
		Keys:            []string{c.GetKey("ds")},
		DataManagerKeys: []string{c.GetKey("dm")},
		TestOnly:        false,
	})
	if err != nil {
		log.Fatalf("RegisterDataSample failed: %v", err)
	}

}

func (c *TestClient) RegisterTrainTask() {
	_, err := c.computeTaskService.RegisterTask(c.ctx, &asset.NewComputeTask{
		Key:            c.GetKey("task"),
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        c.GetKey("algo"),
		ComputePlanKey: c.GetKey("cp"),
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: c.GetKey("dm"),
				DataSampleKeys: []string{c.GetKey("ds")},
			},
		},
	})
	if err != nil {
		log.Fatalf("RegisterComputeTask failed: %v", err)
	}

}

func (c *TestClient) RegisterChildTask() {
	_, err := c.computeTaskService.RegisterTask(c.ctx, &asset.NewComputeTask{
		Key:            c.GetKey("anotherTask"),
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        c.GetKey("algo"),
		ParentTaskKeys: []string{c.GetKey("task")},
		ComputePlanKey: c.GetKey("cp"),
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: c.GetKey("dm"),
				DataSampleKeys: []string{c.GetKey("ds")},
			},
		},
	})
	if err != nil {
		log.Fatalf("RegisterComputeTask failed: %v", err)
	}

}

func (c *TestClient) StartTrainTask() {
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: c.GetKey("task"),
		Action:         asset.ComputeTaskAction_TASK_ACTION_DOING,
	})
	if err != nil {
		log.Fatalf("starting task failed: %v", err)
	}
}

func (c *TestClient) RegisterModel() {
	_, err := c.modelService.RegisterModel(c.ctx, &asset.NewModel{
		ComputeTaskKey: c.GetKey("task"),
		Key:            c.GetKey("model"),
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		Address: &asset.Addressable{
			Checksum:       "5e12e1a2687d81b268558217856547f8a4519f9688933351386a7f902cf1ce5d",
			StorageAddress: "http://somewhere.online/model",
		},
	})
	if err != nil {
		log.Fatalf("RegisterModel failed: %v", err)
	}
}

func (c *TestClient) RegisterComputePlan() {
	_, err := c.computePlanService.RegisterPlan(c.ctx, &asset.NewComputePlan{
		Key: c.GetKey("cp"),
	})
	if err != nil {
		log.Fatalf("RegisterPlan failed: %v", err)
	}
}

func (c *TestClient) QueryComputePlans() {
	resp, err := c.computePlanService.QueryPlans(c.ctx, &asset.QueryPlansParam{})
	if err != nil {
		log.Fatalf("RegisterPlan failed: %v", err)
	}

	c.Plans = resp.Plans
}
