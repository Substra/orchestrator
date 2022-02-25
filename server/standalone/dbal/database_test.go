package dbal

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
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
			db := Database{pool: pool}

			var tx pgx.Tx
			pool.On("BeginTx", mock.MatchedBy(func(ctx context.Context) bool { return true }), tc.txOpts).Return(tx, nil)

			_, err := db.GetTransactionalDBAL(context.Background(), "test", tc.readOnly)

			assert.NoError(t, err)
			pool.AssertExpectations(t)
		})
	}
}