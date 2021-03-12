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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestBlockOnRequest(t *testing.T) {
	limiter := new(ConcurrencyLimiter)

	// Keep track of requests done
	done := 0
	// Fake non-immediate processing
	lock := make(chan interface{})

	var wg sync.WaitGroup

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		defer wg.Done()
		v, ok := req.(int)
		if ok && v == 1 {
			_ = <-lock // Fake processing
		}
		done = v

		return "test", nil
	}

	wg.Add(3)

	// First request come in
	go limiter.Intercept(context.TODO(), 1, unaryInfo, unaryHandler)

	// Second request come in
	go limiter.Intercept(context.TODO(), 2, unaryInfo, unaryHandler)

	// Third request come in
	go limiter.Intercept(context.TODO(), 3, unaryInfo, unaryHandler)

	assert.Equal(t, 0, done, "no request should have been processed")

	lock <- new(struct{}) // first request finishes

	wg.Wait() // Make sure goroutines have been processed
	// The concurrency limiter will make sure that requests are processed one by one REGARDLESS of their order of arrival
	assert.GreaterOrEqual(t, done, 2, "last requests should be done")
}
