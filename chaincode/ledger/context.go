package ledger

import (
	"context"
	"strings"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/common/trace"
	"github.com/owkin/orchestrator/utils"
)

// TransactionContext describes the context passed to every smart contract.
// It's a base contractapi.TransactionContextInterface augmented with ServiceProvider.
type TransactionContext interface {
	contractapi.TransactionContextInterface
	SetContext(context.Context)
	GetContext() context.Context
	GetProvider() (service.DependenciesProvider, error)
	GetDispatcher() event.Dispatcher
	SetRequestID(string)
}

// Context is a TransactionContext implementation
type Context struct {
	context.Context
	contractapi.TransactionContext
	dispatcher event.Dispatcher
}

// NewContext returns a Context instance
func NewContext() *Context {
	// contractapi will NOT use this instance.
	// Instead, it will remember the *type* of the instance, and create a fresh
	// instance from this type. So, don't set properties here: they will not be
	// accessible later.
	return &Context{}
}

// SetContext sets the wrapped context.Context
func (c *Context) SetContext(ctx context.Context) {
	c.Context = ctx
}

// GetContext returns the wrapped context
func (c *Context) GetContext() context.Context {
	return c.Context
}

// GetProvider returns a new instance of ServiceProvider
func (c *Context) GetProvider() (service.DependenciesProvider, error) {
	stub := c.GetStub()

	ctx := c.GetContext()
	db := NewDB(ctx, stub)
	dispatcher := c.GetDispatcher()
	logger := logger.Get(ctx)

	txTimestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return nil, err
	}

	ts := service.NewTimeService(txTimestamp.AsTime())

	return service.NewProvider(logger, db, dispatcher, ts), nil
}

// GetDispatcher returns inner event.Dispatcher
func (c *Context) GetDispatcher() event.Dispatcher {
	if c.dispatcher == nil {
		stub := c.GetStub()
		c.dispatcher = newEventDispatcher(stub)
	}
	return c.dispatcher
}

func (c *Context) SetRequestID(ID string) {
	ctx := context.WithValue(c.Context, trace.RequestIDMarker, ID)
	logger := log.WithField("requestID", ID)
	ctx = log.SetContext(ctx, logger)

	c.Context = ctx
}

type ctxIsInvokeMarker struct{}

var (
	ctxIsEvaluateTransaction = &ctxIsInvokeMarker{}
)

// GetBeforeTransactionHook handles pre-transaction logic:
// - setting the "IsEvaluateTransaction" boolean field.
// Smart contracts MUST use this function as their "BeforeTransaction" attribute
func GetBeforeTransactionHook(contract contractapi.EvaluationContractInterface) func(TransactionContext) error {
	return func(c TransactionContext) error {
		// Determine is method being called is an "Evaluation" method (v.s. "Invoke" method)
		fnName, _ := c.GetStub().GetFunctionAndParameters()
		log.WithField("function", fnName).Debug("Checking if calling function is an eval function")

		evalFuncs := contract.GetEvaluateTransactions()
		isEval := IsEvaluateTransaction(fnName, evalFuncs)

		// Populate context
		ctx := context.WithValue(context.Background(), ctxIsEvaluateTransaction, isEval)
		c.SetContext(ctx)
		return nil
	}
}

// AfterTransactionHook handles post transaction logic:
// - dispatching events
// It MUST be called after orchestration logic happened.
func AfterTransactionHook(ctx TransactionContext, iface interface{}) error {
	return ctx.GetDispatcher().Dispatch()
}

// IsEvaluateTransaction returns true if the passed method name is one of the "evaluate transactions"
// within the evaluateTransactions parameter. The parameter fnName can be either:
// - a method name (eg; "GetAllNodes")
// - a full smart contract + method name (eg; "orchestrator.node:GetAllNodes")
func IsEvaluateTransaction(fnName string, evalFuncs []string) bool {
	idx := strings.LastIndex(fnName, ":")
	if idx != -1 {
		fnName = fnName[idx+1:]
	}
	return utils.StringInSlice(evalFuncs, fnName)
}
