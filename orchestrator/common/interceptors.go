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

package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const headerMSPID = "mspid"

var ignoredMethods = [...]string{
	"grpc.health",
}

// InterceptMSPID is a gRPC interceptor and will make the MSPID from headers available to request context
func InterceptMSPID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("Could not extract metadata")
	}

	if len(md.Get(headerMSPID)) != 1 {
		return nil, fmt.Errorf("Missing or invalid header '%s'", headerMSPID)
	}

	MSPID := md.Get(headerMSPID)[0]

	newCtx := context.WithValue(ctx, ctxMSPIDKey, MSPID)
	return handler(newCtx, req)
}

type ctxMSPIDMarker struct{}

var (
	ctxMSPIDKey = &ctxMSPIDMarker{}
)

// ExtractMSPID retrieves MSPID from request context
// MSPID is expected to be set by InterceptMSPID
func ExtractMSPID(ctx context.Context) (string, error) {
	invocator, ok := ctx.Value(ctxMSPIDKey).(string)
	if !ok {
		return "", errors.New("MSPID not found in context")
	}
	return invocator, nil
}
