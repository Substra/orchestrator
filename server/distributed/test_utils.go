package distributed

import (
	"context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var fabricTimeout = status.New(status.ClientStatus, status.Timeout.ToInt32(), "request timed out or been cancelled", nil)

type mockedInvocator struct {
	mock.Mock
}

func (m *mockedInvocator) Call(ctx context.Context, method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error {
	a := m.Called(ctx, method, param, output)
	return a.Error(0)
}
