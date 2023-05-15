package metrics

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type Task interface {
	fmt.Stringer
	Interval() time.Duration
	Run(ctx context.Context) error
}

// WorkerPool runs tasks at intervals.
type WorkerPool struct {
	tasks   []Task
	workers int
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(tasks []Task, numWorkers int) (*WorkerPool, error) {
	var pool WorkerPool
	pool.workers = numWorkers
	pool.tasks = tasks
	for _, task := range tasks {
		if task.Interval() <= 0 {
			return nil, fmt.Errorf("%s interval must be > 0", task)
		}
	}
	return &pool, nil
}

// Start continuously runs tasks at intervals until the context is canceled.
func (w *WorkerPool) Start(ctx context.Context) {
	ch := make(chan Task)

	var produceGroup errgroup.Group
	for _, task := range w.tasks {
		task := task
		produceGroup.Go(func() error {
			w.produce(ctx, ch, task)
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
	// Two errgroups ensure we do not send on a closed channel. So wait for producers to finish first.
	_ = produceGroup.Wait()
	close(ch)
	_ = workerGroup.Wait()
}

func (w *WorkerPool) produce(ctx context.Context, ch chan<- Task, task Task) {
	submitTask := func() {
		select {
		case <-ctx.Done():
			return
		case ch <- task:
		}
	}

	// Immediately submit the task
	submitTask()

	// Then submit task at interval
	tick := time.NewTicker(task.Interval())
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			submitTask()
		}
	}
}

func (w *WorkerPool) doWork(ctx context.Context, ch <-chan Task) {
	for task := range ch {
		if err := task.Run(ctx); err != nil {
			slog.Warn("Task failed", "task", task.String(), "error", err)
		}
	}
}
