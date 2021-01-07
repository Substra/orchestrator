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

package orchestration

import (
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/mock"
)

// mockDatabaseProvider is a mock implementing DatabaseProvider
type mockServiceProvider struct {
	mock.Mock
}

// GetDatabase will return whatever mock is passed
func (m *mockServiceProvider) GetDatabase() persistence.Database {
	args := m.Called()
	return args.Get(0).(persistence.Database)
}

// GetNodeService will return whatever mock is passed
func (m *mockServiceProvider) GetNodeService() NodeAPI {
	args := m.Called()
	return args.Get(0).(NodeAPI)
}

// GetObjectiveService will return whatever mock is passed
func (m *mockServiceProvider) GetObjectiveService() ObjectiveAPI {
	args := m.Called()
	return args.Get(0).(ObjectiveAPI)
}
