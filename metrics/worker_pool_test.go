package metrics

import (
	"context"
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

	TotalCount *int64
	RunCount   int64
}

func (m *mockTask) String() string {
	return "mock task"
}

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
	return nil
}

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

		pool, err := NewWorkerPool(tasks, 5)
		require.NoError(t, err)

		pool.Start(ctx)

		for _, task := range tasks[:4] {
			require.Equal(t, int64(1), task.(*mockTask).RunCount)
		}
		require.Greater(t, tasks[4].(*mockTask).RunCount, int64(1))
	})

	t.Run("zero duration", func(t *testing.T) {
		tasks := []Task{&mockTask{}}
		_, err := NewWorkerPool(tasks, 1)

		require.Error(t, err)
		require.EqualError(t, err, "mock task interval must be > 0")
	})
}
