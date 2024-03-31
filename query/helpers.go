package query

import (
	"context"
	"github.com/danthegoodman1/QuiCKCRDB/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func ReliableExec(ctx context.Context, pool *pgxpool.Pool, tryTimeout time.Duration, f func(ctx context.Context, q *Queries) error) error {
	return utils.ReliableExec(ctx, pool, tryTimeout, func(ctx context.Context, conn *pgxpool.Conn) error {
		return f(ctx, NewWithTracing(conn))
	})
}

func ReliableExecInSerializedTx(ctx context.Context, pool *pgxpool.Pool, tryTimeout time.Duration, f func(ctx context.Context, q *Queries) error) error {
	ctx, span := createSpan(ctx, "ReliableExecInSerializedTx()")
	defer span.End()
	return utils.ReliableExecInSerializedTx(ctx, pool, tryTimeout, func(ctx context.Context, conn pgx.Tx) error {
		return f(ctx, NewWithTracing(conn))
	})
}

func ReliableExecReadCommittedTx(ctx context.Context, pool *pgxpool.Pool, tryTimeout time.Duration, f func(ctx context.Context, q *Queries) error) error {
	ctx, span := createSpan(ctx, "ReliableExecReadCommitted")
	defer span.End()
	return utils.ReliableExecInReadCommittedTx(ctx, pool, tryTimeout, func(ctx context.Context, conn pgx.Tx) error {
		return f(ctx, NewWithTracing(conn))
	})
}
