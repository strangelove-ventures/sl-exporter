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

	SlashingParamsChain string
	SlashingWindow      float64
}

func (m *mockCosmosMetrics) SetValSlashingParams(chain string, window float64) {
	m.SlashingParamsChain = chain
	m.SlashingWindow = window
}

func (m *mockCosmosMetrics) SetNodeHeight(chain string, height float64) {
	m.NodeHeightChain = chain
	m.NodeHeight = height
}

type mockRestClient struct {
	StubBlock          Block
	StubSlashingParams SlashingParams
}

func (m *mockRestClient) SlashingParams(ctx context.Context) (SlashingParams, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	return m.StubSlashingParams, nil
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

	task := NewRestTask(nil, nil, Chain{Interval: time.Second})
	require.Equal(t, time.Second, task.Interval())

	task = NewRestTask(nil, nil, Chain{})
	require.Equal(t, defaultInterval, task.Interval())
}

func TestRestTask_Run(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		var client mockRestClient
		client.StubBlock.Block.Header.ChainID = "cosmoshub-4"
		client.StubBlock.Block.Header.Height = "1234567890"
		client.StubSlashingParams.Params.SignedBlocksWindow = "10000"

		chain := Chain{
			ChainID: "cosmoshub-4",
			Rest:    []Endpoint{{URL: "http://cosmos.example.com"}, {}},
		}

		var metrics mockCosmosMetrics
		task := NewRestTask(&metrics, &client, chain)

		err := task.Run(ctx)
		require.NoError(t, err)

		require.Equal(t, float64(1234567890), metrics.NodeHeight)
		require.Equal(t, "cosmoshub-4", metrics.NodeHeightChain)

		require.Equal(t, "cosmoshub-4", metrics.SlashingParamsChain)
		require.Equal(t, float64(10000), metrics.SlashingWindow)
	})
}
