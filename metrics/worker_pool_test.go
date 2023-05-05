package metrics

import (
	"context"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var totalCount int64
		const numJobs = 5
		jobs := make([]Job, numJobs)
		for i := 0; i < numJobs; i++ {
			jobs[i] = &mockJob{
				StubInterval: time.Hour,
				CancelAt:     5,
				Cancel:       cancel,
				TotalCount:   &totalCount,
			}
		}

		pool := NewWorkerPool(jobs, 5)
		pool.Start(ctx)
		pool.Wait()

		for _, job := range jobs {
			require.Equal(t, int64(1), job.(*mockJob).RunCount)
		}
	})

	t.Run("variable intervals", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		var totalCount int64
		const numJobs = 5
		jobs := make([]Job, numJobs)
		for i := 0; i < numJobs; i++ {
			jobs[i] = &mockJob{
				StubInterval: time.Duration(r.Intn(100)+1) * time.Microsecond,
				CancelAt:     100,
				Cancel:       cancel,
				TotalCount:   &totalCount,
			}
		}

		pool := NewWorkerPool(jobs, 5)
		pool.Start(ctx)
		pool.Wait()

		for _, job := range jobs {
			require.GreaterOrEqual(t, job.(*mockJob).RunCount, int64(1))
		}
	})
}
