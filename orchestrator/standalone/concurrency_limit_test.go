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
		v, ok := req.(string)
		if ok && v == "wait" {
			_ = <-lock // Fake processing
		}
		done++

		return "test", nil
	}

	wg.Add(2)

	// First request come in
	go limiter.Intercept(context.TODO(), "wait", unaryInfo, unaryHandler)

	// Second request come in
	go limiter.Intercept(context.TODO(), "test", unaryInfo, unaryHandler)

	assert.Equal(t, 0, done, "no request should have been processed")

	lock <- new(struct{}) // first request finishes

	wg.Wait() // Make sure goroutine has been processed
	assert.Equal(t, 2, done, "first request should be done")
}
