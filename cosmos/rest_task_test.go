package cosmos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockCosmosMetrics struct {
	NodeHeightChain string
	NodeHeight      float64
}

func (m *mockCosmosMetrics) SetNodeHeight(chain string, height float64) {
	m.NodeHeightChain = chain
	m.NodeHeight = height
}

type mockRestClient struct {
	StubBlock Block
}

func (m *mockRestClient) LatestBlock(ctx context.Context) (Block, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	return m.StubBlock, nil
}

func TestRestJob_Interval(t *testing.T) {
	t.Parallel()

	job := NewRestJob(nil, nil, Chain{Interval: time.Second})
	require.Equal(t, time.Second, job.Interval())

	job = NewRestJob(nil, nil, Chain{})
	require.Equal(t, defaultInterval, job.Interval())
}

func TestRestJob_String(t *testing.T) {
	t.Parallel()

	job := NewRestJob(nil, nil, Chain{ChainID: "cosmoshub-4"})

	require.Equal(t, "Cosmos REST cosmoshub-4", job.String())
}

func TestRestJob_Run(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		var client mockRestClient

		var blk Block
		blk.Block.Header.Height = "1234567890"
		blk.Block.Header.ChainID = "cosmoshub-4"
		client.StubBlock = blk

		chain := Chain{
			ChainID: "cosmoshub-4",
			Rest:    []Endpoint{{URL: "http://cosmos.example.com"}, {}},
		}

		var metrics mockCosmosMetrics
		job := NewRestJob(&metrics, &client, chain)

		err := job.Run(ctx)
		require.NoError(t, err)

		require.Equal(t, float64(1234567890), metrics.NodeHeight)
		require.Equal(t, "cosmoshub-4", metrics.NodeHeightChain)
	})
}