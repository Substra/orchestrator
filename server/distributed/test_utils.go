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

package distributed

import (
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type mockedInvocator struct {
	mock.Mock
}

func (m *mockedInvocator) Invoke(method string, param protoreflect.ProtoMessage, output protoreflect.ProtoMessage) error {
	a := m.Called(method, param, output)
	return a.Error(0)
}
