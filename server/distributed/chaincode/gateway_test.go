package chaincode

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	req := make(chan chaincodeRequest)
	gw := &Gateway{
		requests: req,
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		request := <-req
		assert.Equal(t, "channel", request.channel)
		assert.Equal(t, "chaincode", request.chaincode)
		assert.Equal(t, "test:Method", request.method)
		assert.Equal(t, []byte("{\"key\":\"test\"}"), request.args)
		wg.Done()
	}()

	gw.Request(
		context.Background(),
		"channel",
		"chaincode",
		"test:Method",
		[]byte("{\"key\":\"test\"}"),
	)

	wg.Wait()
}
