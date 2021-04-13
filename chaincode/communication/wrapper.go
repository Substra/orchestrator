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

// Package communication should be used to communicate with the chaincode.
// This covers invoking smartcontracts and unwrapping response.
package communication

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Wrapper is a wrapper around json representation of protobuf messages.
// It is used to wrap inputs and outputs from smart contracts since contractapi validation is somewhat broken.
// Contractapi expect in/outputs to be serialized in JSON and match their structure type through reflection.
// This goes against our use of protomessages since they are serialized with protojson (not json) and may serialize enums as string instead of int.
// Using json.RawMessage for Param would be more appropriate, unfortunately contractapi has an issue with []byte representation:
// https://jira.hyperledger.org/browse/FABCAG-34
type Wrapper struct {
	Message string `json:"msg"`
}

// Wrap a ProtoMessage in a Wrapper
func Wrap(param protoreflect.ProtoMessage) (*Wrapper, error) {
	p, err := protojson.Marshal(param)
	if err != nil {
		return nil, err
	}

	return &Wrapper{Message: string(p)}, nil
}

// Unwrap a serialized Wrapper into inner ProtoMessage
func Unwrap(data []byte, output protoreflect.ProtoMessage) error {
	var wrappedResponse Wrapper

	err := json.Unmarshal(data, &wrappedResponse)
	if err != nil {
		return err
	}

	err = wrappedResponse.Unwrap(output)
	if err != nil {
		return err
	}

	return nil
}

// Unwrap a JSONPBWrapper into the expected ProtoMessage
func (w *Wrapper) Unwrap(m protoreflect.ProtoMessage) error {
	return protojson.Unmarshal([]byte(w.Message), m)
}
