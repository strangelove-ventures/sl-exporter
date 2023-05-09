package metrics

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/strangelove-ventures/sl-exporter/cosmos"
	"github.com/stretchr/testify/require"
)

type mockValRestClient struct {
	SigningStatusURL     url.URL
	SigningStatusAddress string
	StubSigningStatus    cosmos.SigningStatus
}

func (m *mockValRestClient) SigningStatus(ctx context.Context, baseURL url.URL, consaddress string) (cosmos.SigningStatus, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	m.SigningStatusURL = baseURL
	m.SigningStatusAddress = consaddress
	return m.StubSigningStatus, nil
}

type mockValMetrics struct {
	VailJailChain  string
	ValJailAddress string
	ValJailURL     url.URL
	ValJailStatus  JailStatus
}

func (m *mockValMetrics) SetValJailStatus(chain, consaddress string, restURL url.URL, status JailStatus) {
	m.VailJailChain = chain
	m.ValJailAddress = consaddress
	m.ValJailURL = restURL
	m.ValJailStatus = status
}

func TestCosmosValJob_Interval(t *testing.T) {
	t.Parallel()

	chain := CosmosChain{
		Rest: []Endpoint{
			{URL: "http://cosmos.example.com", Interval: time.Second},
			{URL: "http://another.example.com"},
		},

		Validators: []CosmosValidator{
			{ConsAddress: "cosmosvalcons123"},
			{ConsAddress: "cosmosvalcons567"},
		},
	}

	jobs, err := BuildCosmosValJobs(nil, nil, []CosmosChain{chain})
	require.NoError(t, err)

	require.Len(t, jobs, 4)
	require.Equal(t, time.Second, jobs[0].Interval())
	require.Equal(t, 15*time.Second, jobs[1].Interval())
	require.Equal(t, time.Second, jobs[2].Interval())
	require.Equal(t, 15*time.Second, jobs[3].Interval())
}

func TestCosmosValJob_String(t *testing.T) {
	t.Parallel()

	chain := CosmosChain{
		Rest: []Endpoint{
			{URL: "http://cosmos.example.com", Interval: time.Second},
		},

		Validators: []CosmosValidator{
			{ConsAddress: "cosmosvalcons123"},
			{ConsAddress: "cosmosvalcons567"},
		},
	}
	jobs, err := BuildCosmosValJobs(nil, nil, []CosmosChain{chain})
	require.NoError(t, err)

	require.Len(t, jobs, 2)
	require.Equal(t, "Cosmos validator cosmosvalcons123: http://cosmos.example.com", jobs[0].String())
}

func TestCosmosValJob_Run(t *testing.T) {
	t.Parallel()

	chain := CosmosChain{
		ChainID: "cosmoshub-4",
		Rest: []Endpoint{
			{URL: "http://cosmos.example.com"},
		},

		Validators: []CosmosValidator{
			{ConsAddress: "cosmosvalcons123"},
		},
	}

	now := time.Now()

	t.Run("happy path", func(t *testing.T) {
		for _, tt := range []struct {
			JailedUntil time.Time
			Tombstoned  bool
			WantStatus  JailStatus
		}{
			{time.Time{}, false, JailStatusActive},
			{now.Add(time.Hour), false, JailStatusJailed},
			{time.Time{}, true, JailStatusTombstoned},
			// Tombstoned takes precedence
			{now.Add(time.Hour), true, JailStatusTombstoned},
		} {
			var status cosmos.SigningStatus
			status.ValSigningInfo.Tombstoned = tt.Tombstoned
			status.ValSigningInfo.JailedUntil = tt.JailedUntil
			var client mockValRestClient
			client.StubSigningStatus = status

			var metrics mockValMetrics

			jobs, err := BuildCosmosValJobs(&metrics, &client, []CosmosChain{chain})
			require.NoError(t, err)

			require.Len(t, jobs, 1)
			err = jobs[0].Run(context.Background())

			require.NoError(t, err)
			require.Equal(t, "http://cosmos.example.com", client.SigningStatusURL.String())
			require.Equal(t, client.SigningStatusAddress, "cosmosvalcons123")

			require.Equal(t, "cosmoshub-4", metrics.VailJailChain)
			require.Equal(t, "cosmosvalcons123", metrics.ValJailAddress)
			require.Equal(t, "http://cosmos.example.com", metrics.ValJailURL.String())
			require.Equal(t, tt.WantStatus, metrics.ValJailStatus)
		}
	})

	t.Run("zero state", func(t *testing.T) {
		jobs, err := BuildCosmosValJobs(nil, nil, nil)

		require.NoError(t, err)
		require.Empty(t, jobs)
	})
}
