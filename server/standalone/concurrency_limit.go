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

package standalone

import (
	"context"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

// ConcurrencyLimiter expose a gRPC interceptor method to prevent concurrent requests.
// While counter intuitive, this mechanism is used to reproduce the limitation of the chaincode
// where only one query/invoke can go through at a time.
type ConcurrencyLimiter struct {
	mu sync.Mutex
}

// Intercept a gRPC request, grab the lock and call the underlying handler.
func (cl *ConcurrencyLimiter) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	return handler(ctx, req)
}
