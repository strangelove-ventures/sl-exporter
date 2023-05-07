package metrics

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type Job interface {
	fmt.Stringer
	Interval() time.Duration
	Run(ctx context.Context) error
}

// WorkerPool runs jobs at intervals.
type WorkerPool struct {
	jobs    []Job
	workers int
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(jobs []Job, numWorkers int) (*WorkerPool, error) {
	var pool WorkerPool
	pool.workers = numWorkers
	pool.jobs = jobs
	for _, job := range jobs {
		if job.Interval() <= 0 {
			return nil, fmt.Errorf("%s interval must be > 0", job)
		}
	}
	return &pool, nil
}

// Start continuously runs jobs at intervals until the context is canceled.
func (w *WorkerPool) Start(ctx context.Context) {
	ch := make(chan Job)

	var jobGroup errgroup.Group
	for _, job := range w.jobs {
		job := job
		jobGroup.Go(func() error {
			w.produce(ctx, ch, job)
			return nil
		})
	}

	var workerGroup errgroup.Group
	for i := 0; i < w.workers; i++ {
		workerGroup.Go(func() error {
			w.doWork(ctx, ch)
			return nil
		})
	}
	_ = jobGroup.Wait()
	close(ch)
	_ = workerGroup.Wait()
}

func (w *WorkerPool) produce(ctx context.Context, ch chan<- Job, job Job) {
	submitJob := func() {
		select {
		case <-ctx.Done():
			return
		case ch <- job:
		}
	}

	// Immediately submit the job
	submitJob()

	// Then submit job at interval
	tick := time.NewTicker(job.Interval())
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			submitJob()
		}
	}
}

func (w *WorkerPool) doWork(ctx context.Context, ch <-chan Job) {
	for job := range ch {
		if err := job.Run(ctx); err != nil {
			slog.Warn("Job failed", "job", job.String(), "error", err)
		}
	}
}
