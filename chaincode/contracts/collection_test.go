package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEvaluateMethod(t *testing.T) {
	p := NewContractCollection()
	assert.Equal(t, true, p.IsEvaluateMethod("orchestrator.objective:QueryObjectives"))
	assert.Equal(t, false, p.IsEvaluateMethod("orchestrator.objective:RegisterObjective"))
	assert.Equal(t, false, p.IsEvaluateMethod("orchestrator.objective:DoesntExist"))
}
