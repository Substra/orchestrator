// +build e2e

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

package e2e

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
)

func (c *AppClient) registerNode(t *testing.T) {
	client := asset.NewNodeServiceClient(c.conn)
	_, err := client.RegisterNode(c.ctx, &asset.NodeRegistrationParam{})
	if err != nil {
		t.Errorf("RegisterNode failed: %v", err)
	}
}

func (c *AppClient) registerAlgo(t *testing.T) {
	algoClient := asset.NewAlgoServiceClient(c.conn)
	_, err := algoClient.RegisterAlgo(c.ctx, &asset.NewAlgo{
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
		t.Errorf("RegisterAlgo failed: %v", err)
	}

}

func (c *AppClient) registerDataManager(t *testing.T) {
	dataManagerClient := asset.NewDataManagerServiceClient(c.conn)
	_, err := dataManagerClient.RegisterDataManager(c.ctx, &asset.NewDataManager{
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
		t.Errorf("RegisterDataManager failed: %v", err)
	}

}

func (c *AppClient) registerDataSample(t *testing.T) {
	dataSampleClient := asset.NewDataSampleServiceClient(c.conn)
	_, err := dataSampleClient.RegisterDataSample(c.ctx, &asset.NewDataSample{
		Keys:            []string{c.GetKey("ds")},
		DataManagerKeys: []string{c.GetKey("dm")},
		TestOnly:        false,
	})
	if err != nil {
		t.Errorf("RegisterDataSample failed: %v", err)
	}

}

func (c *AppClient) registerTrainTask(t *testing.T) {
	computeTaskClient := asset.NewComputeTaskServiceClient(c.conn)
	_, err := computeTaskClient.RegisterTask(c.ctx, &asset.NewComputeTask{
		Key:            c.GetKey("task"),
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        c.GetKey("algo"),
		Rank:           0,
		ComputePlanKey: c.GetKey("cp"),
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: c.GetKey("dm"),
				DataSampleKeys: []string{c.GetKey("ds")},
			},
		},
	})
	if err != nil {
		t.Errorf("RegisterComputeTask failed: %v", err)
	}

}

func (c *AppClient) registerChildTask(t *testing.T) {
	computeTaskClient := asset.NewComputeTaskServiceClient(c.conn)
	_, err := computeTaskClient.RegisterTask(c.ctx, &asset.NewComputeTask{
		Key:            c.GetKey("anotherTask"),
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        c.GetKey("algo"),
		Rank:           1,
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
		t.Errorf("RegisterComputeTask failed: %v", err)
	}

}

func (c *AppClient) startTrainTask(t *testing.T) {
	computeTaskClient := asset.NewComputeTaskServiceClient(c.conn)
	_, err := computeTaskClient.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: c.GetKey("task"),
		Action:         asset.ComputeTaskAction_TASK_ACTION_DOING,
	})
	if err != nil {
		t.Errorf("starting task failed: %v", err)
	}
}

func (c *AppClient) registerModel(t *testing.T) {
	modelClient := asset.NewModelServiceClient(c.conn)
	_, err := modelClient.RegisterModel(c.ctx, &asset.NewModel{
		ComputeTaskKey: c.GetKey("task"),
		Key:            c.GetKey("model"),
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		Address: &asset.Addressable{
			Checksum:       "5e12e1a2687d81b268558217856547f8a4519f9688933351386a7f902cf1ce5d",
			StorageAddress: "http://somewhere.online/model",
		},
	})
	if err != nil {
		t.Errorf("RegisterModel failed: %v", err)
	}
}
