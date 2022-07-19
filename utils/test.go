package utils

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/mock"
)

// AnyContext will match any context.Context: empty ones as well as WithValue ones.
var AnyContext = mock.MatchedBy(func(c context.Context) bool {
	// if the passed in parameter does not implement the context.Context interface, the
	// wrapping MatchedBy will panic - so we can simply return true, since we
	// know it's a context.Context if execution flow makes it here.
	return true
})

// Tx will generate a mock for pgx.Tx
type Tx interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) (err error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	LargeObjects() pgx.LargeObjects
	Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error)
	Conn() *pgx.Conn
}

// MockConn augments pgxmock.PgxConnIface with a WaitForNotification method so that
// it implements dbal.Conn interface.
type MockConn struct {
	pgxmock.PgxConnIface
	mock.Mock
}

func (m *MockConn) WaitForNotification(ctx context.Context) (*pgconn.Notification, error) {
	ret := m.Called(ctx)

	var r0 *pgconn.Notification
	if rf, ok := ret.Get(0).(func(context.Context) *pgconn.Notification); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pgconn.Notification)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewMockConn() (*MockConn, error) {
	conn, err := pgxmock.NewConn()
	if err != nil {
		return nil, err
	}

	return &MockConn{PgxConnIface: conn}, nil
}
