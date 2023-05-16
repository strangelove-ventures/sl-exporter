package metrics

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

type mockTask struct {
	Cancel       context.CancelFunc
	CancelAt     int64
	StubInterval time.Duration
	StubErr      error

	TotalCount *int64
	RunCount   int64
}

func (m *mockTask) Group() string { return "mock" }
func (m *mockTask) ID() string    { return "my_task" }

func (m *mockTask) Interval() time.Duration {
	return m.StubInterval
}

func (m *mockTask) Run(ctx context.Context) error {
	if ctx == nil {
		panic("nil context")
	}
	atomic.AddInt64(&m.RunCount, 1)
	atomic.AddInt64(m.TotalCount, 1)
	if atomic.LoadInt64(m.TotalCount) >= m.CancelAt {
		m.Cancel()
	}
	return m.StubErr
}

type mockTaskErrorMetrics func(group string)

func (fn mockTaskErrorMetrics) IncFailedTask(group string) { fn(group) }

func TestWorkerPool(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var totalCount int64
		tasks := make([]Task, 4)
		for i := 0; i < 4; i++ {
			tasks[i] = &mockTask{
				StubInterval: time.Hour,
				CancelAt:     10,
				Cancel:       cancel,
				TotalCount:   &totalCount,
			}
		}

		tasks = append(tasks, &mockTask{
			StubInterval: time.Millisecond,
			CancelAt:     10,
			Cancel:       cancel,
			TotalCount:   &totalCount,
		})

		pool, err := NewWorkerPool(tasks, 5, nil)
		require.NoError(t, err)

		pool.Start(ctx)

		for _, task := range tasks[:4] {
			require.Equal(t, int64(1), task.(*mockTask).RunCount)
		}
		require.Greater(t, tasks[4].(*mockTask).RunCount, int64(1))
	})

	t.Run("zero duration", func(t *testing.T) {
		tasks := []Task{&mockTask{}}
		_, err := NewWorkerPool(tasks, 1, nil)

		require.Error(t, err)
		require.EqualError(t, err, "mock:my_task interval must be > 0")
	})

	t.Run("task error", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var totalCount int64
		tasks := make([]Task, 2)
		for i := 0; i < 2; i++ {
			tasks[i] = &mockTask{
				StubInterval: time.Hour,
				CancelAt:     2,
				Cancel:       cancel,
				TotalCount:   &totalCount,
				StubErr:      fmt.Errorf("boom"),
			}
		}

		var metricCount int64
		metrics := mockTaskErrorMetrics(func(group string) {
			if group != "mock" {
				panic(fmt.Errorf("unexpected group: %s", group))
			}
			atomic.AddInt64(&metricCount, 1)
		})

		pool, err := NewWorkerPool(tasks, 5, metrics)
		require.NoError(t, err)

		pool.Start(ctx)
		require.Equal(t, int64(2), metricCount)
	})
}
