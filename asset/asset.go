// Copyright 2018-2020 Owkin, inc.
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

package asset

import "fmt"

// AssetType is use to check the type of an asset
type AssetType uint8

// Const representing the types of asset findable in the ledger
const (
	ObjectiveType AssetType = iota
	DataManagerType
	DataSampleType
	AlgoType
	CompositeAlgoType
	AggregateAlgoType
	TraintupleType
	CompositeTraintupleType
	AggregatetupleType
	TesttupleType
	ComputePlanType
	// when adding a new type here, don't forget to update
	// the String() function in utils.go
)

type Asset struct {
	Type        AssetType `json:"asset_type"`
	Permissions []string
}

// ChecksumAddress stores a checksum and a Storage Address
type ChecksumAddress struct {
	Checksum       string `json:"checksum"`
	StorageAddress string `json:"storage_address"`
}

// String returns a string representation for an asset type
func (assetType AssetType) String() string {
	switch assetType {
	case ObjectiveType:
		return "objective"
	case DataManagerType:
		return "data_manager"
	case DataSampleType:
		return "data_sample"
	case AlgoType:
		return "algo"
	case CompositeAlgoType:
		return "composite_algo"
	case AggregateAlgoType:
		return "aggregate_algo"
	case TraintupleType:
		return "traintuple"
	case CompositeTraintupleType:
		return "composite_traintuple"
	case AggregatetupleType:
		return "aggregatetuple"
	case TesttupleType:
		return "testtuple"
	case ComputePlanType:
		return "compute_plan"
	default:
		return fmt.Sprintf("(unknown asset type: %d)", assetType)
	}
}
