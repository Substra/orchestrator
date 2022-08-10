package chaincode

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/utils"
)

type mockChaincodeRequester struct {
	mock.Mock
}

func (m *mockChaincodeRequester) Request(ctx context.Context, channel string, chaincode string, method string, arguments []byte) (<-chan []byte, <-chan error) {
	args := m.Called(ctx, channel, chaincode, method, arguments)
	return args.Get(0).(<-chan []byte), args.Get(1).(<-chan error)
}

func TestContractInvocator(t *testing.T) {
	requester := &mockChaincodeRequester{}
	invocator := NewContractInvocator(requester, "channel", "chaincode")

	assert.Implementsf(t, (*Invocator)(nil), invocator, "ContractInvocator should implements Invocator")
}

func TestParamWrapping(t *testing.T) {
	requester := &mockChaincodeRequester{}
	invocator := NewContractInvocator(requester, "channel", "chaincode")

	// Invocation param is a protoreflect.ProtoMessage
	param := &asset.QueryAlgosParam{PageToken: "uuid", PageSize: 20}
	wrapper, err := communication.Wrap(context.Background(), param)
	require.NoError(t, err)

	// Which is serialized
	serializedInput, err := json.Marshal(wrapper)
	require.NoError(t, err)

	// Response is also a wrapper
	response := &asset.QueryAlgosResponse{Algos: []*asset.Algo{}, NextPageToken: "test"}
	wrappedResponse, err := communication.Wrap(context.Background(), response)
	require.NoError(t, err)
	// Then serialized to match contractapi
	serializedResponse, err := json.Marshal(wrappedResponse)
	require.NoError(t, err)

	resChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		resChan <- serializedResponse
	}()

	requester.On("Request", utils.AnyContext, "channel", "chaincode", "orchestrator.algo:QueryAlgos", serializedInput).
		Once().
		Return((<-chan []byte)(resChan), (<-chan error)(errChan))

	output := &asset.QueryAlgosResponse{}
	err = invocator.Call(context.TODO(), "orchestrator.algo:QueryAlgos", param, output)
	assert.NoError(t, err)

	assert.Equal(t, "test", output.NextPageToken, "response should be properly unwrapped")
}

func TestNoOutput(t *testing.T) {
	requester := &mockChaincodeRequester{}
	invocator := NewContractInvocator(requester, "channel", "chaincode")

	expectedInput := getEmptyExpectedInput(t)

	resChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		resChan <- nil
	}()

	requester.On("Request", utils.AnyContext, "channel", "chaincode", "org.test:NoOutput", expectedInput).
		Once().
		Return((<-chan []byte)(resChan), (<-chan error)(errChan))

	err := invocator.Call(context.TODO(), "org.test:NoOutput", nil, nil)
	assert.NoError(t, err)
}

func getEmptyExpectedInput(t *testing.T) []byte {
	// Invocation param is a protoreflect.ProtoMessage
	wrapper, err := communication.Wrap(context.Background(), nil)
	require.NoError(t, err)

	// Which is serialized
	serializedInput, err := json.Marshal(wrapper)
	require.NoError(t, err)

	return serializedInput
}
