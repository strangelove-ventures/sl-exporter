package metrics

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/strangelove-ventures/sl-exporter/rest"
	"github.com/stretchr/testify/require"
)

type mockCosmosMetrics struct {
	NodeHeightChain  string
	NodeHeightRPCURL url.URL
	NodeHeight       float64
}

func (m *mockCosmosMetrics) SetNodeHeight(chain string, rpcURL url.URL, height float64) {
	m.NodeHeightChain = chain
	m.NodeHeightRPCURL = rpcURL
	m.NodeHeight = height
}

type mockRPCClient struct {
	StubBlock rest.Block
	StatusURL url.URL
}

func (m *mockRPCClient) LatestBlock(ctx context.Context, baseURL url.URL) (rest.Block, error) {
	if ctx == nil {
		panic("nil context")
	}
	m.StatusURL = baseURL
	return m.StubBlock, nil
}

func TestCosmosRestJob_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		var metrics mockCosmosMetrics
		var client mockRPCClient
		client.StubBlock.Block.Header.Height = "1234567890"

		chains := []CosmosChain{
			{
				ChainID: "cosmoshub-4",
				Rest:    []Endpoint{{URL: "http://rpc.example.com", Interval: time.Second}, {}},
			},
			{
				ChainID: "akash",
				Rest:    []Endpoint{{}},
			},
		}

		jobs, err := BuildCosmosRestJobs(&metrics, &client, chains)
		require.NoError(t, err)

		require.Len(t, jobs, 3)

		job := jobs[0]

		require.Equal(t, "Cosmos REST http://rpc.example.com", job.String())
		require.Equal(t, time.Second, job.Interval())

		err = job.Run(ctx)
		require.NoError(t, err)

		wantURL := url.URL{Scheme: "http", Host: "rpc.example.com"}
		require.Equal(t, wantURL, client.StatusURL)

		require.Equal(t, wantURL, metrics.NodeHeightRPCURL)
		require.Equal(t, float64(1234567890), metrics.NodeHeight)
		require.Equal(t, "cosmoshub-4", metrics.NodeHeightChain)

		job = jobs[2]
		err = job.Run(ctx)
		require.NoError(t, err)
		require.Equal(t, "akash", metrics.NodeHeightChain)
	})

	t.Run("default interval", func(t *testing.T) {
		var metrics mockCosmosMetrics
		var client mockRPCClient

		chains := []CosmosChain{
			{
				ChainID: "akash",
				Rest:    []Endpoint{{}},
			},
		}
		job, err := BuildCosmosRestJobs(&metrics, &client, chains)
		require.NoError(t, err)

		require.Equal(t, 5*time.Second, job[0].Interval())
	})
}
