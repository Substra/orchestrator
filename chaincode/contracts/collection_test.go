package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEvaluateMethod(t *testing.T) {
	p := NewContractCollection()
	assert.True(t, p.IsEvaluateMethod("orchestrator.algo:QueryAlgos"))
	assert.False(t, p.IsEvaluateMethod("orchestrator.algo:RegisterAlgo"))
	assert.False(t, p.IsEvaluateMethod("orchestrator.algo:DoesntExist"))
}
