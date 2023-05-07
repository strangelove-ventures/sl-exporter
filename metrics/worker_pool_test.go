package metrics

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

type mockJob struct {
	Cancel       context.CancelFunc
	CancelAt     int64
	StubInterval time.Duration

	TotalCount *int64
	RunCount   int64
}

func (m *mockJob) String() string {
	return "mock job"
}

func (m *mockJob) Interval() time.Duration {
	return m.StubInterval
}

func (m *mockJob) Run(ctx context.Context) error {
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
		jobs := make([]Job, 4)
		for i := 0; i < 4; i++ {
			jobs[i] = &mockJob{
				StubInterval: time.Hour,
				CancelAt:     10,
				Cancel:       cancel,
				TotalCount:   &totalCount,
			}
		}

		jobs = append(jobs, &mockJob{
			StubInterval: time.Millisecond,
			CancelAt:     10,
			Cancel:       cancel,
			TotalCount:   &totalCount,
		})

		pool, err := NewWorkerPool(jobs, 5)
		require.NoError(t, err)

		pool.Start(ctx)

		for _, job := range jobs[:4] {
			require.Equal(t, int64(1), job.(*mockJob).RunCount)
		}
		require.Greater(t, jobs[4].(*mockJob).RunCount, int64(1))
	})

	t.Run("zero duration", func(t *testing.T) {
		jobs := []Job{&mockJob{}}
		_, err := NewWorkerPool(jobs, 1)

		require.Error(t, err)
		require.EqualError(t, err, "mock job interval must be > 0")
	})
}
