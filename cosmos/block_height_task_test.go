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

func TestRestTask_Interval(t *testing.T) {
	t.Parallel()

	task := NewBlockHeightTask(nil, nil, Chain{Interval: time.Second})
	require.Equal(t, time.Second, task.Interval())

	task = NewBlockHeightTask(nil, nil, Chain{})
	require.Equal(t, defaultInterval, task.Interval())
}

func TestRestTask_Run(t *testing.T) {
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
		task := NewBlockHeightTask(&metrics, &client, chain)

		err := task.Run(ctx)
		require.NoError(t, err)

		require.Equal(t, float64(1234567890), metrics.NodeHeight)
		require.Equal(t, "cosmoshub-4", metrics.NodeHeightChain)
	})
}
