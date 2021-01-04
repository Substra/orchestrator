// Copyright 2020 Owkin Inc.
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

package dataset

import (
	"errors"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract manages Datasets
type SmartContract struct {
	contractapi.Contract
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	return &SmartContract{}
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryDataManager", "QueryDataManagers", "QueryDataset", "QuerySamples"}
}

// RegisterDataManager stores a new dataManager in the ledger
func (s *SmartContract) RegisterDataManager() error {
	return errors.New("unimplemented")
}

// RegisterSample stores new Sample in the ledger (one or more).
func (s *SmartContract) RegisterSample() error {
	return errors.New("unimplemented")
}

// UpdateSample associates one or more dataManagerKeys to one or more Samples
func (s *SmartContract) UpdateSample() error {
	return errors.New("unimplemented")
}

// UpdateDataManager associates a objectiveKey to an existing dataManager
func (s *SmartContract) UpdateDataManager() error {
	return errors.New("unimplemented")
}

// QueryDataManager returns dataManager and its key
func (s *SmartContract) QueryDataManager() error {
	return errors.New("unimplemented")
}

// QueryDataManagers returns all DataManagers of the ledger
func (s *SmartContract) QueryDataManagers() error {
	return errors.New("unimplemented")
}

// QueryDataset returns info about a dataManager and all related dataSample
func (s *SmartContract) QueryDataset() error {
	return errors.New("unimplemented")
}

// QuerySamples list samples
func (s *SmartContract) QuerySamples() error {
	return errors.New("unimplemented")
}
