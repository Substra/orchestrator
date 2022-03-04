package utils

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// AnyContext will match any context.Context: empty ones as well as WithValue ones.
var AnyContext = mock.MatchedBy(func(c context.Context) bool {
	// if the passed in parameter does not implement the context.Context interface, the
	// wrapping MatchedBy will panic - so we can simply return true, since we
	// know it's a context.Context if execution flow makes it here.
	return true
})
