package chaincode

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/chaincode/communication"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Invocator describe how to invoke chaincode in a somewhat generic way.
// This is the Gandalf of fabric.
type Invocator interface {
	Call(ctx context.Context, method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error
}

// ContractInvocator implements the Invocator interface.
type ContractInvocator struct {
	requester Requester
	channel   string
	chaincode string
}

// NewContractInvocator creates an Invocator based on given smart contract.
func NewContractInvocator(requester Requester, channel, chaincode string) *ContractInvocator {
	return &ContractInvocator{requester, channel, chaincode}
}

// Call will evaluate or invoke a transaction to the ledger, deserializing its result in the output parameter.
// The choice of evaluation or invocation is based on contracts.AllEvaluateTransactions.
func (i *ContractInvocator) Call(ctx context.Context, method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error {
	var err error

	logger := log.Ctx(ctx).With().Str("method", method).Interface("param", param).Logger()
	start := time.Now()

	wrapper, err := communication.Wrap(ctx, param)
	if err != nil {
		return err
	}
	args, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	outChan, errChan := i.requester.Request(ctx, i.channel, i.chaincode, method, args)

	select {
	case data := <-outChan:
		if output != nil {
			err := communication.Unwrap(data, output)
			if err != nil {
				return err
			}
			elapsed := time.Since(start)

			logger.Debug().Dur("duration", elapsed).Msg("Successfully called chaincode")

			return nil
		}
	case err = <-errChan:
		return err
	case <-ctx.Done():
		logger.Info().Msg("context done before invocation response")
		return ctx.Err()
	}

	return err
}
