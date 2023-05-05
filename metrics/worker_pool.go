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
	pool.workers = numWorkers
	pool.jobs = trackJobs
	return &pool
}

// Start continuously runs jobs at intervals until the context is canceled.
func (w *WorkerPool) Start(ctx context.Context) {
	var wg sync.WaitGroup
	ch := make(chan Job)
	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go w.doWork(ctx, ch, &wg)
	}

	for {
		for _, job := range w.jobs {
			select {
			case <-ctx.Done():
				close(ch)
				wg.Wait()
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

func (w *WorkerPool) doWork(ctx context.Context, ch <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range ch {
		if err := job.Run(ctx); err != nil {
			logrus.WithError(err).WithField("job", job.String()).Error("Job failed")
		}
	}
}
