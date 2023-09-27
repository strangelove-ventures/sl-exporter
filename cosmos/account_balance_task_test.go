package cosmos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockAccountClient struct {
	GotAddress  string
	GotDenom    string
	StubBalance AccountBalance
}

func (m *mockAccountClient) AccountBalance(ctx context.Context, address, denom string) (AccountBalance, error) {
	if ctx == nil {
		panic("nil context")
	}
	m.GotAddress = address
	m.GotDenom = denom
	return m.StubBalance, nil
}

type mockAccountMetrics struct {
	GotChain   string
	GotAlias   string
	GotAddress string
	GotDenom   string
	GotBalance float64
}

func (m *mockAccountMetrics) SetAccountBalance(chain, alias, address, denom string, balance float64) {
	m.GotChain = chain
	m.GotAlias = alias
	m.GotAddress = address
	m.GotDenom = denom
	m.GotBalance = balance
}

func TestAccountTask_Run(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		t.Run("happy path", func(t *testing.T) {
			chain := Chain{
				ChainID: "osmosis-1",
				Accounts: []Account{
					{
						Address: "osmo1234",
						Alias:   "osmosis",
						Denoms:  []string{"uosmo", "ibc/ABC"},
					},
					{
						Address: "osmo456",
						Alias:   "osmosis2",
						Denoms:  []string{"uosmo"},
					},
				},
			}

			var client mockAccountClient
			client.StubBalance = AccountBalance{
				Account: "osmo1234",
				Denom:   "uosmo",
				Amount:  1234567890,
			}
			var metrics mockAccountMetrics

			tasks := NewAccountTasks(&metrics, &client, chain)

			require.Equal(t, 3, len(tasks))

			task := tasks[0]

			err := task.Run(ctx)
			require.NoError(t, err)

			require.Equal(t, "osmo1234", client.GotAddress)
			require.Equal(t, "uosmo", client.GotDenom)

			require.Equal(t, "osmosis-1", metrics.GotChain)
			require.Equal(t, "osmosis", metrics.GotAlias)
			require.Equal(t, "osmo1234", metrics.GotAddress)
			require.Equal(t, "uosmo", metrics.GotDenom)
			require.Equal(t, 1234567890.0, metrics.GotBalance)
		})
	})
}
