package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/strangelove-ventures/sl-exporter/cosmos"
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
	StubBlock cosmos.Block
}

func (m *mockRestClient) LatestBlock(ctx context.Context) (cosmos.Block, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	return m.StubBlock, nil
}

func TestCosmosRestJob_Interval(t *testing.T) {
	t.Parallel()

	chains := []CosmosChain{
		{Interval: time.Second},
		{},
	}

	jobs := BuildCosmosRestJobs(nil, nil, chains)

	require.Len(t, jobs, 2)
	require.Equal(t, time.Second, jobs[0].Interval())
	require.Equal(t, 15*time.Second, jobs[1].Interval())
}

func TestCosmosRestJob_String(t *testing.T) {
	t.Parallel()

	chains := []CosmosChain{
		{ChainID: "cosmoshub-4"},
	}

	jobs := BuildCosmosRestJobs(nil, nil, chains)

	require.Equal(t, "Cosmos REST cosmoshub-4", jobs[0].String())
}

func TestCosmosRestJob_Run(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		var client mockRestClient

		var blk cosmos.Block
		blk.Block.Header.Height = "1234567890"
		blk.Block.Header.ChainID = "cosmoshub-4"
		client.StubBlock = blk

		chains := []CosmosChain{
			{
				ChainID: "cosmoshub-4",
				Rest:    []Endpoint{{URL: "http://cosmos.example.com"}, {}},
			},
			{
				ChainID: "akash-1234",
				Rest:    []Endpoint{{URL: "http://akash.example.com"}},
			},
		}

		var metrics mockCosmosMetrics
		jobs := BuildCosmosRestJobs(&metrics, &client, chains)

		require.Len(t, jobs, 2)

		job := jobs[0]

		err := job.Run(ctx)
		require.NoError(t, err)

		require.Equal(t, float64(1234567890), metrics.NodeHeight)
		require.Equal(t, "cosmoshub-4", metrics.NodeHeightChain)

		job = jobs[1]
		err = job.Run(ctx)
		require.NoError(t, err)
		require.Equal(t, "akash-1234", metrics.NodeHeightChain)
	})
}
