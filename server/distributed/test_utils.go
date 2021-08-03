package distributed

import (
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type mockedInvocator struct {
	mock.Mock
}

func (m *mockedInvocator) Call(ctx context.Context, method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error {
	a := m.Called(ctx, method, param, output)
	return a.Error(0)
}

// AnyContext will match any context.Context: empty ones as well as WithValue ones.
var AnyContext = mock.MatchedBy(func(c context.Context) bool {
	// if the passed in parameter does not implement the context.Context interface, the
	// wrapping MatchedBy will panic - so we can simply return true, since we
	// know it's a context.Context if execution flow makes it here.
	return true
})
