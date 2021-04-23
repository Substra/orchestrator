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

package standalone

import (
	"github.com/owkin/orchestrator/lib/asset"
)

// GetDataset implements persistence.DatasetDBAL
func (d *DBAL) GetDataset(id string) (*asset.Dataset, error) {

	datamanager, err := d.GetDataManager(id)

	if err != nil {
		return nil, err
	}

	trainDataSampleKeys, err := d.GetDataSamplesKeysByDataManager(id, false)
	if err != nil {
		return nil, err
	}

	testDataSampleKeys, err := d.GetDataSamplesKeysByDataManager(id, true)
	if err != nil {
		return nil, err
	}

	dataset := &asset.Dataset{
		DataManager:         datamanager,
		TrainDataSampleKeys: trainDataSampleKeys,
		TestDataSampleKeys:  testDataSampleKeys,
	}

	return dataset, nil
}
