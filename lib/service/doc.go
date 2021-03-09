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

// Package service holds the business logic allowing the manipulation of assets.
// The main entry point is the ServiceProvider, from which the services are accessible.
// Services can rely on each other when multiple assets are involved in the same transaction.
// Storage is isolated in the persistence layer and services should only deal with their dedicated asset's persistence.
package service
