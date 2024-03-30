package quickcrdb

import "time"

type (
	QueueItem struct {
		QueueZone   string
		ID          int64
		Payload     string
		VestingTime time.Time
	}
)
