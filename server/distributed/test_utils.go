package distributed

import (
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type mockedInvocator struct {
	mock.Mock
}

func (m *mockedInvocator) Call(method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error {
	a := m.Called(method, param, output)
	return a.Error(0)
}
