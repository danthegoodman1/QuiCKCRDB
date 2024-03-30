package quickcrdb

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"runtime"
	"time"
)

type (
	Worker struct {
		pool   *pgxpool.Pool
		config *workerConfig
	}
	workerConfig struct {
		managerRoutines      int
		workerRoutines       int
		sequential           bool
		peekMax              int
		selectionFrac        float64
		selectionMax         int
		processingBound      int
		dequeueMax           time.Duration
		pointerLeaseDuration time.Duration
		// min time a queue remains  empty before its pointer is deleted
		pointerMinInactive time.Duration
	}
)

var (
	defaultConfig = &workerConfig{
		managerRoutines: runtime.NumCPU(),
		workerRoutines:  runtime.NumCPU(),
		sequential:      false,
		peekMax:         100,
		selectionFrac:   0.1,
		selectionMax:    10,
		processingBound: runtime.NumCPU(),
	}
)

func NewWorker(pool *pgxpool.Pool, opts ...WorkerOption) (*Worker, error) {
	worker := &Worker{
		pool:   pool,
		config: defaultConfig,
	}

	for _, opt := range opts {
		opt(worker.config)
	}

	return worker, nil
}
