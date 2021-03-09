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

package ledger

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/service"
)

// TransactionContext describes the context passed to every smart contract.
// It's a base TransactionContext augmented with ServiceProvider.
type TransactionContext interface {
	contractapi.TransactionContextInterface
	GetProvider() service.DependenciesProvider
	GetDispatcher() event.Dispatcher
}

// Context is a TransactionContext implementation
type Context struct {
	contractapi.TransactionContext
	dispatcher event.Dispatcher
}

// ensureDispatcher instanciates an eventDispatcher if needed.
// This kind of lazy loading is needed since the stub in not available at instanciation time.
func (c *Context) ensureDispatcher() {
	if c.dispatcher == nil {
		stub := c.GetStub()
		dispatcher := newEventDispatcher(stub)
		c.dispatcher = dispatcher
	}
}

// GetProvider returns a new instance of ServiceProvider
func (c *Context) GetProvider() service.DependenciesProvider {
	stub := c.GetStub()
	db := NewDB(stub)
	c.ensureDispatcher()

	return service.NewProvider(db, c.dispatcher)
}

// GetDispatcher returns inner event.Dispatcher
func (c *Context) GetDispatcher() event.Dispatcher {
	c.ensureDispatcher()
	return c.dispatcher
}

// NewContext returns a Context instance
func NewContext() *Context {
	return &Context{}
}

// AfterTransactionHook handles post transaction logic:
// - dispatching events
// It MUST be called after orchestration logic happened.
func AfterTransactionHook(ctx TransactionContext, iface interface{}) error {
	return ctx.GetDispatcher().Dispatch()
}
