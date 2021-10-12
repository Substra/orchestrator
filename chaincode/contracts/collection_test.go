package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEvaluateMethod(t *testing.T) {
	p := NewContractCollection()
	assert.Equal(t, true, p.IsEvaluateMethod("orchestrator.metric:QueryMetrics"))
	assert.Equal(t, false, p.IsEvaluateMethod("orchestrator.metric:RegisterMetric"))
	assert.Equal(t, false, p.IsEvaluateMethod("orchestrator.metric:DoesntExist"))
}
