package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/UltimateTournament/backoff/v4"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"os"
	"strconv"
	"time"
)

func getEnvOrDefault(env, defaultVal string) string {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		return e
	}
}

func getEnvOrDefaultInt(env string, defaultVal int64) int64 {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		intVal, err := strconv.ParseInt(e, 10, 64)
		if err != nil {
			return defaultVal
		}

		return intVal
	}
}

// ReliableExec wrapper exists so caller stack skipping works
func ReliableExec(ctx context.Context, pool *pgxpool.Pool, tryTimeout time.Duration, f func(ctx context.Context, conn *pgxpool.Conn) error) error {
	err := reliableExec(ctx, pool, tryTimeout, func(ctx context.Context, conn *pgxpool.Conn) error {
		return f(ctx, conn)
	})
	if err != nil {
		return err
	}
	return nil
}

func ReliableExecInSerializedTx(ctx context.Context, pool *pgxpool.Pool, tryTimeout time.Duration, f func(ctx context.Context, conn pgx.Tx) error) error {
	err := reliableExec(ctx, pool, tryTimeout, func(ctx context.Context, conn *pgxpool.Conn) error {
		return crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
			return f(ctx, tx)
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func ReliableExecInReadCommittedTx(ctx context.Context, pool *pgxpool.Pool, tryTimeout time.Duration, f func(ctx context.Context, conn pgx.Tx) error) error {
	err := reliableExec(ctx, pool, tryTimeout, func(ctx context.Context, conn *pgxpool.Conn) error {
		return crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL READ COMMITTED")
			if err != nil {
				return fmt.Errorf("failed to set read committed: %w", err)
			}
			return f(ctx, tx)
		})
	})
	if err != nil {
		return err
	}
	return nil
}

var dbTracer = otel.GetTracerProvider().Tracer("db")

func reliableExec(ctx context.Context, pool *pgxpool.Pool, tryTimeout time.Duration, f func(ctx context.Context, conn *pgxpool.Conn) error) error {
	cfg := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)

	return backoff.Retry(func() error {
		tryCtx, cancel := context.WithTimeout(ctx, tryTimeout)
		defer cancel()
		acquireCtx, acquireSpan := dbTracer.Start(tryCtx, "pool.Acquire")
		conn, err := pool.Acquire(acquireCtx)
		acquireSpan.End()
		if err != nil {
			err = fmt.Errorf("failed pool.Acquire: %w", err)
			// for DeadlineExceeded we must check the outer context, as the inner one is expected to fail due to `tryTimeout`
			if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return backoff.Permanent(err)
			}
			return err
		}
		defer conn.Release()
		funcCtx, funcSpan := dbTracer.Start(tryCtx, "execWithCon")
		err = f(funcCtx, conn)
		funcSpan.End()
		if IsPermSQLErr(err) {
			return backoff.Permanent(err)
		}
		// for DeadlineExceeded we must check the outer context, as the inner one is expected to fail due to `tryTimeout`
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return backoff.Permanent(err)
		}
		return err
	}, cfg)
}

func IsPermSQLErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			// This is a duplicate key - unique constraint
			return true
		}
		if pgErr.Code == "42703" {
			// Column does not exist
			return true
		}
	}
	return false
}
