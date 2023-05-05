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

func ToJobs[T Job](jobs []T) []Job {
	result := make([]Job, len(jobs))
	for i := range jobs {
		result[i] = jobs[i]
	}
	return result
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
	var pool WorkerPool
	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)
	}
	pool.workers = numWorkers
	pool.jobs = trackJobs
	return &pool
}

// Start continuously runs jobs at intervals until the context is canceled.
func (w *WorkerPool) Start(ctx context.Context) {
	ch := make(chan Job)
	for i := 0; i < w.workers; i++ {
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

// Wait blocks until Start's context is cancelled. Callers should wait to ensure all goroutines exit.
func (w *WorkerPool) Wait() { w.wg.Wait() }
