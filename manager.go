package quickcrdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/danthegoodman1/QuiCKCRDB/query"
	"time"
)

func (w *Worker) managerObtainTopLevelQueue(ctx context.Context, queue query.QuickTopLevelQueue) error {
	leaseID := "" // TODO: generate uuid
	// TODO: make obtain timeout customizable
	err := query.ReliableExecInSerializedTx(ctx, w.pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		_, err = q.ObtainTopLevelQueue(ctx, query.ObtainTopLevelQueueParams{
			LeaseID: sql.NullString{
				Valid:  true,
				String: leaseID,
			},
			VestingTime: time.Now().Add(w.queueZoneLeaseDuration),
			QueueZone:   queue.QueueZone,
		})
		if err != nil {
			return fmt.Errorf("error in ObtainTopLevelQueue: %w", err)
		}

		return
	})
	if err != nil {
		return err
	}

	// TODO: if obtained, dequeue messages and send to worker threads

	return nil
}
