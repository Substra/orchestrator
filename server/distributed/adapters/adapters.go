package adapters

import (
	"context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/substra/orchestrator/server/common/interceptors"
)

// isFabricTimeoutRetry will return true if we are in a retry and the last error was a fabric timeout
func isFabricTimeoutRetry(ctx context.Context) bool {
	prevErr := interceptors.GetLastError(ctx)
	if prevErr == nil {
		return false
	}

	st, ok := status.FromError(prevErr)
	return ok && st.Code == int32(status.Timeout)
}
