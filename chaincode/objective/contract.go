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

package objective

import (
	"errors"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/ledger"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/assets/dataset"
	objectiveAsset "github.com/owkin/orchestrator/lib/assets/objective"
)

func getServiceFromContext(ctx contractapi.TransactionContextInterface) (objectiveAsset.API, error) {
	db, err := ledger.GetLedgerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return objectiveAsset.NewService(db), nil
}

// SmartContract manages objectives
type SmartContract struct {
	contractapi.Contract
	serviceFactory func(contractapi.TransactionContextInterface) (objectiveAsset.API, error)
}

// NewSmartContract creates a smart contract to be used in a chaincode
func NewSmartContract() *SmartContract {
	return &SmartContract{
		serviceFactory: getServiceFromContext,
	}
}

// RegisterObjective creates a new objective in world state
// If the key exists, it will override the existing value with the new one
func (s *SmartContract) RegisterObjective(
	ctx contractapi.TransactionContextInterface,
	key string,
	name string,
	description *assets.Addressable,
	metricsName string,
	metrics *assets.Addressable,
	testDataset *dataset.Dataset,
	metadata map[string]string,
	permissions *assets.Permissions,
) error {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return err
	}

	// TODO: validation (description/metrics/etc)

	o := objectiveAsset.Objective{
		Key:         key,
		Name:        name,
		Description: description,
		MetricsName: metricsName,
		Metrics:     metrics,
		Metadata:    metadata,
		Permissions: permissions,
	}

	// TODO: add Dataset ???

	err = service.RegisterObjective(&o)
	return err
}

// QueryObjectives returns the objectives
func (s *SmartContract) QueryObjectives(ctx contractapi.TransactionContextInterface) ([]*objectiveAsset.Objective, error) {
	service, err := s.serviceFactory(ctx)
	if err != nil {
		return nil, err
	}

	return service.GetObjectives()
}

// QueryLeaderboard returns for an objective all its certified testtuples with a done status
func (s *SmartContract) QueryLeaderboard(ctx contractapi.TransactionContextInterface, key string, sortOrder assets.SortOrder) (*objectiveAsset.Leaderboard, error) {
	return nil, errors.New("unimplemented")
}

// GetEvaluateTransactions returns functions of SmartContract not to be tagged as submit
func (s *SmartContract) GetEvaluateTransactions() []string {
	return []string{"QueryObjectives", "QueryLeaderboard"}
}
