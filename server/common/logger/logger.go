package logger

import (
	"context"

	"github.com/go-playground/log/v7"
)

// Get returns the logger stored in the context, or a new Default log Entry if none is found.
// See also: `AddLogger(ctx, req, info, handler)`.
func Get(ctx context.Context) log.Entry {
	return log.GetContext(ctx)
}
