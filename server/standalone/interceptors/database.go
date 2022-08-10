package interceptors

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/substra/orchestrator/server/common"
	"google.golang.org/grpc"
)

type DatabaseConnInterceptor struct {
	dbURL string
}

func NewDatabaseConnInterceptor(dbURL string) *DatabaseConnInterceptor {
	return &DatabaseConnInterceptor{dbURL: dbURL}
}

// StreamServerInterceptor will make a new database connection
// available from the context of each request
func (i *DatabaseConnInterceptor) StreamServerInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := stream.Context()

	conn, err := pgx.Connect(ctx, i.dbURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	newCtx := WithDatabaseConn(ctx, conn)
	streamWithContext := common.BindStreamToContext(newCtx, stream)
	return handler(srv, streamWithContext)
}

type ctxDatabaseConnInterceptorMarker struct{}

var ctxDatabaseConnKey = &ctxDatabaseConnInterceptorMarker{}

func WithDatabaseConn(ctx context.Context, conn *pgx.Conn) context.Context {
	return context.WithValue(ctx, ctxDatabaseConnKey, conn)
}

// ExtractDatabaseConn will return the *pgx.Conn injected in context
func ExtractDatabaseConn(ctx context.Context) (*pgx.Conn, error) {
	conn, ok := ctx.Value(ctxDatabaseConnKey).(*pgx.Conn)
	if !ok {
		return nil, errors.New("database connection not found in context")
	}
	return conn, nil
}
