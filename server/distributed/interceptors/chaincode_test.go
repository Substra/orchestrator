package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/server/distributed/chaincode"
)

func TestExtractInvocator(t *testing.T) {
	ctx := context.TODO()

	i := &chaincode.MockInvocator{}
	ctxWithInvocator := WithInvocator(ctx, i)

	extracted, err := ExtractInvocator(ctxWithInvocator)
	assert.NoError(t, err, "extraction should not fail")
	assert.Equal(t, i, extracted, "Invocator should be extracted from context")

	_, err = ExtractInvocator(ctx)
	assert.Error(t, err, "Extraction should fail on empty context")
}
