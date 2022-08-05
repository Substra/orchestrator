// Package communication should be used to communicate with the chaincode.
// This covers invoking smartcontracts and unwrapping response.
package communication

import (
	"context"
	"encoding/json"

	"github.com/owkin/orchestrator/server/common/interceptors"
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
	Message   string `json:"msg"`
	RequestID string `json:"request_id"`
}

// Wrap a ProtoMessage in a Wrapper
func Wrap(ctx context.Context, param protoreflect.ProtoMessage) (*Wrapper, error) {
	p, err := protojson.Marshal(param)
	if err != nil {
		return nil, err
	}

	wrapper := &Wrapper{
		Message:   string(p),
		RequestID: interceptors.GetRequestID(ctx),
	}

	return wrapper, nil
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
