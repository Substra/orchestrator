package dataset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateTransactions(t *testing.T) {
	contract := NewSmartContract()

	queries := []string{
		"GetDataset",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
