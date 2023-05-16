package metrics

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type Task interface {
	// Group returns the group this task belongs to.
	Group() string
	// ID is a unique identifier for this task.
	ID() string
	// Interval is how often this task runs.
	Interval() time.Duration
	Run(ctx context.Context) error
}

type TaskErrorMetrics interface {
	IncFailedTask(group string)
}

// WorkerPool runs tasks at intervals.
type WorkerPool struct {
	tasks   []Task
	metrics TaskErrorMetrics
	workers int
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(tasks []Task, numWorkers int, metrics TaskErrorMetrics) (*WorkerPool, error) {
	for _, task := range tasks {
		if task.Interval() <= 0 {
			return nil, fmt.Errorf("%s:%s interval must be > 0", task.Group(), task.ID())
		}
	}
	return &WorkerPool{
		tasks:   tasks,
		metrics: metrics,
		workers: numWorkers,
	}, nil
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
			w.metrics.IncFailedTask(task.Group())
			slog.Error("Task failed", "group", task.Group(), "error", err)
		}
	}
}
