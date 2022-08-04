package ledger

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-playground/log/v7"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/common"
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
}

// Context is a TransactionContext implementation
type Context struct {
	context.Context
	contractapi.TransactionContext
	queue      event.Queue
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
	logger := logger.Get(ctx)

	txTimestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return nil, err
	}

	ts := service.NewTimeService(txTimestamp.AsTime())

	return service.NewProvider(logger, db, c.getQueue(), ts, stub.GetChannelID()), nil
}

func (c *Context) getQueue() event.Queue {
	if c.queue == nil {
		c.queue = new(common.MemoryQueue)
	}
	return c.queue
}

// GetDispatcher returns inner event.Dispatcher
func (c *Context) GetDispatcher() event.Dispatcher {
	if c.dispatcher == nil {
		stub := c.GetStub()
		c.dispatcher = newEventDispatcher(stub, c.getQueue(), logger.Get(c.Context))
	}
	return c.dispatcher
}

type ctxIsInvokeMarker struct{}

var (
	ctxIsEvaluateTransaction = &ctxIsInvokeMarker{}
)

// GetBeforeTransactionHook handles pre-transaction logic:
// - setting the "IsEvaluateTransaction" boolean field;
// - populating context with logger and requestID;
// Smart contracts MUST use this function as their "BeforeTransaction" attribute.
// The requestID is automatically extracted from call parameters, as long as the first parameter is a communication.Wrapper.
func GetBeforeTransactionHook(contract contractapi.EvaluationContractInterface) func(TransactionContext) error {
	return func(c TransactionContext) error {
		// Determine is method being called is an "Evaluation" method (v.s. "Invoke" method)
		fnName, args := c.GetStub().GetFunctionAndParameters()

		l := log.WithField("function", fnName)

		var requestID string
		if len(args) > 0 {
			w := new(communication.Wrapper)
			if err := json.Unmarshal([]byte(args[0]), &w); err == nil {
				requestID = w.RequestID
			} else {
				l.WithError(err).Warn("cannot extract request ID")
			}
		}

		evalFuncs := contract.GetEvaluateTransactions()
		isEval := IsEvaluateTransaction(fnName, evalFuncs)

		l = l.WithField("isEval", isEval).WithField("requestID", requestID)

		// Populate context
		ctx := context.WithValue(context.Background(), ctxIsEvaluateTransaction, isEval)
		ctx = context.WithValue(ctx, trace.RequestIDMarker, requestID)
		ctx = log.SetContext(ctx, l)
		c.SetContext(ctx)

		l.Debug("transaction context initialized")
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
// - a method name (eg; "GetAllOrganizations")
// - a full smart contract + method name (eg; "orchestrator.organization:GetAllOrganizations")
func IsEvaluateTransaction(fnName string, evalFuncs []string) bool {
	idx := strings.LastIndex(fnName, ":")
	if idx != -1 {
		fnName = fnName[idx+1:]
	}
	return utils.SliceContains(evalFuncs, fnName)
}
