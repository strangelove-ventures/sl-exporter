package cosmos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockParamsMetrics struct {
	GotSlashingChain  string
	GotSlashingWindow float64
}

func (m *mockParamsMetrics) SetValSlashingParams(chain string, window float64) {
	m.GotSlashingChain = chain
	m.GotSlashingWindow = window
}

type mockParamsClient struct {
	StubSlashingParams SlashingParams
}

func (m mockParamsClient) SlashingParams(ctx context.Context) (SlashingParams, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	return m.StubSlashingParams, nil
}

func TestValParamsTask_Interval(t *testing.T) {
	t.Parallel()

	task := NewValParamsTask(nil, nil, Chain{Interval: time.Second})

	require.Equal(t, 5*time.Minute, task.Interval())
}

func TestValParamsTask_Run(t *testing.T) {
	var metrics mockParamsMetrics
	var client mockParamsClient
	client.StubSlashingParams.Params.SignedBlocksWindow = "10000"

	task := NewValParamsTask(&metrics, client, Chain{ChainID: "cosmoshub-4"})

	err := task.Run(context.Background())
	require.NoError(t, err)

	require.Equal(t, "cosmoshub-4", metrics.GotSlashingChain)
	require.Equal(t, float64(10000), metrics.GotSlashingWindow)
}
