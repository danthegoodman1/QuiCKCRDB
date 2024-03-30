package quickcrdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

type quickTopLevelQueue struct {
	QueueZone   string
	VestingTime time.Time
	LeaseID     sql.NullString
	HashToken   int64
}

func selectVestedTopLevelQueues(ctx context.Context, conn pgx.Tx, token int) ([]quickTopLevelQueue, error) {
	rows, err := conn.Query(ctx, `select *
from quick_top_level_queue
where hash_token < $1
and vesting_time <= now()`, token)
	if err != nil {
		return nil, fmt.Errorf("error selecting top level queue: %w", err)
	}
	defer rows.Close()
	var results []quickTopLevelQueue
	for rows.Next() {
		var row quickTopLevelQueue
		err = rows.Scan(
			&row.QueueZone,
			&row.VestingTime,
			&row.LeaseID,
			&row.HashToken,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error completing row read: %w", err)
	}
	return results, nil
}
