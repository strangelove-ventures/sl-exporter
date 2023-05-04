package metrics

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockCosmosMetrics struct {
	NodeHeightChain string
	NodeHeightURL   *url.URL
	NodeHeight      float64
}

func (m *mockCosmosMetrics) SetNodeHeight(chain string, rpcURL *url.URL, height float64) {
	m.NodeHeightChain = chain
	m.NodeHeightURL = rpcURL
	m.NodeHeight = height
}

func TestRPCPoller_Poll(t *testing.T) {
	t.Parallel()

	chains := []CosmosChain{
		{
			Chain: "cosmoshub-4",
			RPCs: []RPC{
				{URL: "http://1.example.com"},
				{URL: "http://2.example.com"},
			},
		},
		{
			Chain: "akash",
			RPCs: []RPC{
				{URL: "http://3.example.com"},
			},
		},
	}

	poller, err := NewRPCPoller(nil, chains)
	require.NoError(t, err)

	poller.Poll(nil)
}
