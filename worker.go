package quickcrdb

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"runtime"
	"sync/atomic"
	"time"
)

type (
	Worker struct {
		pool         *pgxpool.Pool
		config       *workerConfig
		hashRingSize int
		stopScanner  chan any
		stopManagers chan any
		stopWorkers  chan any
		shuttingDown *atomic.Bool
	}
	workerConfig struct {
		managerRoutines      int
		workerRoutines       int
		sequential           bool
		peekMax              int
		selectionFrac        float64
		selectionMax         int
		processingBound      int
		dequeueMax           int
		pointerLeaseDuration time.Duration
		// min time a queue remains empty before its pointer is deleted
		pointerMinInactive          time.Duration
		vestingTimeRewriteThreshold time.Duration
		fifo                        bool
	}
)

var (
	defaultConfig = &workerConfig{
		managerRoutines:             runtime.NumCPU(),
		workerRoutines:              runtime.NumCPU(),
		sequential:                  false,
		peekMax:                     100,
		selectionFrac:               0.1,
		selectionMax:                10,
		processingBound:             runtime.NumCPU(),
		dequeueMax:                  10,
		pointerLeaseDuration:        time.Second,
		pointerMinInactive:          time.Second * 30,
		vestingTimeRewriteThreshold: time.Millisecond * 250,
	}
)

func NewWorker(pool *pgxpool.Pool, hashRingSize int, opts ...WorkerOption) (*Worker, error) {
	worker := &Worker{
		pool:         pool,
		config:       defaultConfig,
		hashRingSize: hashRingSize,
		shuttingDown: &atomic.Bool{},
	}

	for _, opt := range opts {
		opt(worker.config)
	}

	worker.stopScanner = make(chan any, 1)
	worker.stopManagers = make(chan any, worker.config.managerRoutines)
	worker.stopWorkers = make(chan any, worker.config.workerRoutines)

	return worker, nil
}

func (w *Worker) scanner() {
	token := 0
	for {
		select {
		case <-w.stopScanner:
			logger.Info().Msg("scanner exiting")
			return
		default:
			// TODO: Scan hash token for queue zones
		}

		token++
		if token > w.hashRingSize {
			token = 0
		}
	}
}

func (w *Worker) manager() {

}

func (w *Worker) worker() {

}

// StopScanner tells the scanner goroutine. It is safe to crash all goroutines, so on exit you don't even need to stop
func (w *Worker) StopScanner() {
	if w.shuttingDown.CompareAndSwap(false, true) {
		w.stopScanner <- nil
		for i := 0; i < w.config.managerRoutines; i++ {
			w.stopManagers <- nil
		}
		for i := 0; i < w.config.workerRoutines; i++ {
			w.stopWorkers <- nil
		}
	}
}
