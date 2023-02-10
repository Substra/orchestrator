package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEvaluateMethod(t *testing.T) {
	p := NewContractCollection()
	assert.True(t, p.IsEvaluateMethod("orchestrator.function:QueryFunctions"))
	assert.False(t, p.IsEvaluateMethod("orchestrator.function:RegisterFunction"))
	assert.False(t, p.IsEvaluateMethod("orchestrator.function:DoesntExist"))
}
