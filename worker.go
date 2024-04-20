package quickcrdb

import (
	"context"
	"fmt"
	"github.com/danthegoodman1/QuiCKCRDB/query"
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

		scannerInterval time.Duration
		scannerTicker   *time.Ticker

		queueZoneLeaseDuration time.Duration
		queueItemLeaseDuration time.Duration

		managerRecv chan query.QuickTopLevelQueue
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
		scannerInterval             time.Duration
		managerRecvBuffer           int
	}

	// WorkerFunc is invoked by each worker thread when it receives and item for processing
	WorkerFunc func(ctx context.Context, item QueueItem) error
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
		scannerInterval:             time.Millisecond * 100,
		managerRecvBuffer:           100,
	}
)

func NewWorker(pool *pgxpool.Pool, hashRingSize int, queueZoneLeaseDuration, queueItemLeaseDuration time.Duration, workerFunction WorkerFunc, opts ...WorkerOption) (*Worker, error) {
	worker := &Worker{
		pool:                   pool,
		config:                 defaultConfig,
		hashRingSize:           hashRingSize,
		shuttingDown:           &atomic.Bool{},
		queueItemLeaseDuration: queueItemLeaseDuration,
		queueZoneLeaseDuration: queueZoneLeaseDuration,
	}

	for _, opt := range opts {
		opt(worker.config)
	}

	worker.stopScanner = make(chan any, 1)
	worker.stopManagers = make(chan any, worker.config.managerRoutines)
	worker.stopWorkers = make(chan any, worker.config.workerRoutines)

	worker.scannerTicker = time.NewTicker(worker.config.scannerInterval)

	worker.managerRecv = make(chan query.QuickTopLevelQueue, worker.config.managerRecvBuffer)

	go worker.launchScanner()

	return worker, nil
}

func (w *Worker) launchScanner() {
	token := 0
	for {
		select {
		case <-w.stopScanner:
			logger.Info().Msg("launchScanner exiting")
			return
		case <-w.scannerTicker.C:
			// Scan hash token for queue zones
			w.scanHashToken(token)
		}

		token++
		if token > w.hashRingSize {
			token = 0
		}
	}
}

func (w *Worker) scanHashToken(token int) {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.scannerInterval)
	defer cancel()

	// Get queue zones
	var topLevelQueues []query.QuickTopLevelQueue
	err := query.ReliableExecReadCommittedTx(ctx, w.pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		topLevelQueues, err = q.SelectTopLevelQueues(ctx, int64(token))
		if err != nil {
			return fmt.Errorf("error in selectVestedTopLevelQueues: %w", err)
		}

		return
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("error scanning top level queues")
		return
	}

	// TODO: Send queue zone pointers to manager
}

func (w *Worker) launchManager(managerID string) {
	for {
		select {
		case <-w.stopManagers:
			logger.Info().Msgf("manager %s exiting", managerID)
			return
		case queue := <-w.managerRecv:
			err := w.managerObtainTopLevelQueue(context.Background(), queue) // timeout in function
			if err != nil {
				// Crash
				logger.Fatal().Err(err).Msg("error in managerObtainTopLevelQueue")
			}
		}
	}
}

func (w *Worker) launchWorker(workerID string) {
	// TODO: listen for dequeued message from the manager
	// TODO: process messages
}

// StopScanner tells the launchScanner goroutine. It is safe to crash all goroutines, so on exit you don't even need to stop
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
