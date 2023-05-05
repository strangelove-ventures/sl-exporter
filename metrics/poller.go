package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Job interface {
	fmt.Stringer
	Interval() time.Duration
	Run(ctx context.Context) error
}

// WorkerPool runs jobs at intervals.
type WorkerPool struct {
	jobs    []*workerJob
	workers int
	wg      sync.WaitGroup
}

type workerJob struct {
	Job
	lastRun time.Time
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(jobs []Job, numWorkers int) *WorkerPool {
	trackJobs := make([]*workerJob, len(jobs))
	for i := range jobs {
		trackJobs[i] = &workerJob{Job: jobs[i]}
	}
	return &WorkerPool{
		workers: numWorkers,
		jobs:    trackJobs,
	}
}

// Do continuously runs jobs at intervals until the context is canceled.
func (w *WorkerPool) Do(ctx context.Context) {
	ch := make(chan Job)
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.doWork(ctx, ch)
	}

	for {
		for _, job := range w.jobs {
			select {
			case <-ctx.Done():
				close(ch)
				return
			default:
				if time.Since(job.lastRun) < job.Interval() {
					continue
				}
				ch <- job
				job.lastRun = time.Now()
			}
		}
	}
}

func (w *WorkerPool) doWork(ctx context.Context, ch <-chan Job) {
	defer w.wg.Done()
	for job := range ch {
		if err := job.Run(ctx); err != nil {
			logrus.WithError(err).WithField("job", job.String()).Error("Job failed")
		}
	}
}

// Wait blocks until Do's context is cancelled. Callers should wait to ensure all goroutines exit.
func (w *WorkerPool) Wait() { w.wg.Wait() }
