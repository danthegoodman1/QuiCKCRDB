package quickcrdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/danthegoodman1/QuiCKCRDB/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

func (w *Worker) managerObtainTopLevelQueue(ctx context.Context, queue query.QuickTopLevelQueue) error {
	leaseUUID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("error in NewUUID: %w", err)
	}

	leaseID := leaseUUID.String()

	// TODO: make obtain timeout customizable
	err = query.ReliableExecInSerializedTx(ctx, w.pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		_, err = q.ObtainTopLevelQueue(ctx, query.ObtainTopLevelQueueParams{
			NewLease: sql.NullString{
				Valid:  true,
				String: leaseID,
			},
			VestingTime: time.Now().Add(w.queueZoneLeaseDuration),
			QueueZone:   queue.QueueZone,
			KnownLease:  queue.LeaseID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// We didn't obtain it
				logger.Debug().Msgf("failed to obtain queue zone '%s', (someone else probably obtained it first)")
				return nil
			}
			return fmt.Errorf("error in ObtainTopLevelQueue: %w", err)
		}

		return
	})
	if err != nil {
		return err
	}

	// We obtained it

	// TODO: Check if it's empty

	// Dequeue messages and send to worker threads
	err = query.ReliableExecInSerializedTx(ctx, w.pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		_, err = q.DequeueItems(ctx, query.DequeueItemsParams{
			HashToken: queue.HashToken,
			QueueZone: queue.QueueZone,
			Limit:     int32(w.config.dequeueMax),
			VestingTime: sql.NullTime{
				Valid: true,
				Time:  time.Now().Add(w.queueItemLeaseDuration),
			},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// We didn't obtain it
				logger.Debug().Msgf("failed to obtain queue item")
				return nil
			}
			return fmt.Errorf("error in DequeueItems: %w", err)
		}

		return
	})
	if err != nil {
		return err
	}

	// TODO: get min vesting time

	// TODO: check if we need to update pointer index vesting time

	// TODO: check if we need to delete pointer p

	return nil
}
