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
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestRegisterComputeTask(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(app.getDialer), grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	ctx = metadata.AppendToOutgoingContext(ctx, "mspid", "MyOrg1MSP", "channel", "test-register-compute-task")

	algoKey := "673afa7e-11e7-4539-98d6-a11acafb9dbd"
	dataManagerKey := "347020bc-0f74-470b-b241-21b1e0c368d6"

	nodeClient := asset.NewNodeServiceClient(conn)
	_, err = nodeClient.RegisterNode(ctx, &asset.NodeRegistrationParam{})
	if err != nil {
		t.Errorf("RegisterNode failed: %v", err)
	}

	algoClient := asset.NewAlgoServiceClient(conn)
	_, err = algoClient.RegisterAlgo(ctx, &asset.NewAlgo{
		Key:      algoKey,
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

	dataManagerClient := asset.NewDataManagerServiceClient(conn)
	_, err = dataManagerClient.RegisterDataManager(ctx, &asset.NewDataManager{
		Key:            dataManagerKey,
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

	dataSampleClient := asset.NewDataSampleServiceClient(conn)
	_, err = dataSampleClient.RegisterDataSample(ctx, &asset.NewDataSample{
		Keys:            []string{"fef7e71a-27aa-47de-bd9b-44e86d063af8"},
		DataManagerKeys: []string{dataManagerKey},
		TestOnly:        false,
	})
	if err != nil {
		t.Errorf("RegisterDataManager failed: %v", err)
	}

	computeTaskClient := asset.NewComputeTaskServiceClient(conn)
	_, err = computeTaskClient.RegisterTask(ctx, &asset.NewComputeTask{
		Key:            "0c742318-bac2-43b7-8ce9-f629567d930c",
		Category:       asset.ComputeTaskCategory_TASK_TRAIN,
		AlgoKey:        algoKey,
		Rank:           0,
		ComputePlanKey: "a26c95ef-7283-44f6-a280-5a3e7df1ee86",
		Data: &asset.NewComputeTask_Train{
			Train: &asset.NewTrainTaskData{
				DataManagerKey: dataManagerKey,
				DataSampleKeys: []string{"fef7e71a-27aa-47de-bd9b-44e86d063af8"},
			},
		},
	})
	if err != nil {
		t.Errorf("RegisterDataManager failed: %v", err)
	}
}
