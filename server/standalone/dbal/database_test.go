package dbal

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/utils"
)

func TestTransactionIsolation(t *testing.T) {
	cases := map[string]struct {
		readOnly bool
		txOpts   pgx.TxOptions
	}{
		"read-only transaction options": {
			true,
			pgx.TxOptions{
				IsoLevel:   pgx.ReadCommitted,
				AccessMode: pgx.ReadOnly,
			},
		},
		"read-write transaction options": {
			false,
			pgx.TxOptions{
				IsoLevel: pgx.Serializable,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			pool := new(MockPgPool)
			db := Database{Pool: pool}

			tx := new(utils.MockTx)
			pool.On("BeginTx", utils.AnyContext, tc.txOpts).Return(tx, nil)

			_, err := db.BeginTransaction(context.Background(), tc.readOnly)
			assert.NoError(t, err)
			tx.AssertExpectations(t)
			pool.AssertExpectations(t)
		})
	}
}
