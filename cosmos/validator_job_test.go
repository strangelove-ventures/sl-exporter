package cosmos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockValRestClient struct {
	SigningStatusAddress string
	StubSigningStatus    SigningStatus
}

func (m *mockValRestClient) SigningStatus(ctx context.Context, consaddress string) (SigningStatus, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	m.SigningStatusAddress = consaddress
	return m.StubSigningStatus, nil
}

type mockValMetrics struct {
	VailJailChain  string
	ValJailAddress string
	ValJailStatus  JailStatus
}

func (m *mockValMetrics) SetValJailStatus(chain, consaddress string, status JailStatus) {
	m.VailJailChain = chain
	m.ValJailAddress = consaddress
	m.ValJailStatus = status
}

func TestValidatorJob_Interval(t *testing.T) {
	t.Parallel()

	chains := []Chain{
		{Interval: time.Second, Validators: []Validator{{ConsAddress: "1"}, {ConsAddress: "2"}}},
		{Validators: []Validator{{ConsAddress: "3"}}}, // empty chain
	}

	jobs := BuildValidatorJobs(nil, nil, chains)

	require.Len(t, jobs, 3)
	require.Equal(t, time.Second, jobs[0].Interval())
	require.Equal(t, time.Second, jobs[1].Interval())
	require.Equal(t, defaultInterval, jobs[2].Interval())
}

func TestValidatorJob_String(t *testing.T) {
	t.Parallel()

	chain := Chain{
		ChainID: "cosmoshub-4",
		Rest: []Endpoint{
			{URL: "http://cosmos.example.com"},
		},

		Validators: []Validator{
			{ConsAddress: "cosmosvalcons123"},
			{ConsAddress: "cosmosvalcons567"},
		},
	}
	jobs := BuildValidatorJobs(nil, nil, []Chain{chain})

	require.Len(t, jobs, 2)
	require.Equal(t, "Cosmos validator cosmosvalcons123: cosmoshub-4", jobs[0].String())
}

func TestValdatorJob_Run(t *testing.T) {
	t.Parallel()

	chain := Chain{
		ChainID: "cosmoshub-4",
		Rest: []Endpoint{
			{URL: "http://cosmos.example.com"},
		},

		Validators: []Validator{
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
			var status SigningStatus
			status.ValSigningInfo.Tombstoned = tt.Tombstoned
			status.ValSigningInfo.JailedUntil = tt.JailedUntil
			var client mockValRestClient
			client.StubSigningStatus = status

			var metrics mockValMetrics

			jobs := BuildValidatorJobs(&metrics, &client, []Chain{chain})

			require.Len(t, jobs, 1)
			err := jobs[0].Run(context.Background())

			require.NoError(t, err)
			require.Equal(t, client.SigningStatusAddress, "cosmosvalcons123")

			require.Equal(t, "cosmoshub-4", metrics.VailJailChain)
			require.Equal(t, "cosmosvalcons123", metrics.ValJailAddress)
			require.Equal(t, tt.WantStatus, metrics.ValJailStatus)
		}
	})

	t.Run("zero state", func(t *testing.T) {
		jobs := BuildValidatorJobs(nil, nil, nil)

		require.Empty(t, jobs)
	})
}
