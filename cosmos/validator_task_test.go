package cosmos

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockValRestClient struct {
	StubBlock Block

	SigningInfoAddress string
	StubSigningInfo    SigningInfo
}

func (m *mockValRestClient) LatestBlock(ctx context.Context) (Block, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	return m.StubBlock, nil
}

func (m *mockValRestClient) SigningInfo(ctx context.Context, consaddress string) (SigningInfo, error) {
	_, ok := ctx.Deadline()
	if !ok {
		panic("expected deadline in context")
	}
	m.SigningInfoAddress = consaddress
	return m.StubSigningInfo, nil
}

type mockValMetrics struct {
	GotChain        string
	GotAddr         string
	GotJailStatus   JailStatus
	GotSignedBlock  float64
	GotMissedBlocks float64

	SignedBlockCount int
}

func (m *mockValMetrics) SetValJailStatus(chain, consaddress string, status JailStatus) {
	m.GotChain = chain
	m.GotAddr = consaddress
	m.GotJailStatus = status
}

func (m *mockValMetrics) IncValSignedBlocks(chain, consaddress string) {
	m.SignedBlockCount++
	m.GotChain = chain
	m.GotAddr = consaddress
}

func (m *mockValMetrics) SetValSignedBlock(chain, consaddress string, height float64) {
	m.GotChain = chain
	m.GotAddr = consaddress
	m.GotSignedBlock = height
}

func (m *mockValMetrics) SetValMissedBlocks(chain, consaddress string, missed float64) {
	m.GotChain = chain
	m.GotAddr = consaddress
	m.GotMissedBlocks = missed
}

func TestValidatorTask_Interval(t *testing.T) {
	t.Parallel()

	chain := Chain{Interval: time.Second, Validators: []Validator{{ConsAddress: "1"}, {ConsAddress: "2"}}}

	tasks := BuildValidatorTasks(nil, nil, chain)

	require.Len(t, tasks, 2)
	require.Equal(t, time.Second, tasks[0].Interval())
	require.Equal(t, time.Second, tasks[1].Interval())

	chain = Chain{Validators: []Validator{{ConsAddress: "1"}}}
	tasks = BuildValidatorTasks(nil, nil, chain)

	require.Len(t, tasks, 1)
	require.Equal(t, defaultInterval, tasks[0].Interval())
}

func TestValidatorTask_Run(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const addr = `cosmosvalcons164q2kq3q3psj436t9p7swmdlh39rw73wpy6qx6`

	t.Run("zero state", func(t *testing.T) {
		tasks := BuildValidatorTasks(nil, nil, Chain{})

		require.Empty(t, tasks)
	})

	t.Run("happy path - signed blocks", func(t *testing.T) {
		chain := Chain{
			ChainID: "cosmoshub-4",
			Validators: []Validator{
				{ConsAddress: addr},
			},
		}

		client := new(mockValRestClient)
		var metrics mockValMetrics
		tasks := BuildValidatorTasks(&metrics, client, chain)
		client.StubBlock.Block.LastCommit.Height = "1"
		client.StubSigningInfo.ValSigningInfo.MissedBlocksCounter = "0"

		require.Len(t, tasks, 1)
		task := tasks[0]
		err := task.Run(ctx)
		require.NoError(t, err)

		require.Zero(t, metrics.SignedBlockCount)
		require.Zero(t, metrics.GotSignedBlock)

		var block Block
		require.NoError(t, json.Unmarshal(latestBlockFixture, &block))
		block.Block.LastCommit.Height = "9001"
		client.StubBlock = block

		err = task.Run(ctx)
		require.NoError(t, err)

		require.Equal(t, 1, metrics.SignedBlockCount)
		require.Equal(t, "cosmoshub-4", metrics.GotChain)
		require.Equal(t, addr, metrics.GotAddr)

		require.Equal(t, float64(9001), metrics.GotSignedBlock)
	})

	t.Run("happy path - jail status", func(t *testing.T) {
		now := time.Now()

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
			var status SigningInfo
			status.ValSigningInfo.Tombstoned = tt.Tombstoned
			status.ValSigningInfo.JailedUntil = tt.JailedUntil
			status.ValSigningInfo.MissedBlocksCounter = "0"

			var client mockValRestClient
			client.StubSigningInfo = status
			client.StubBlock.Block.LastCommit.Height = "1"

			var metrics mockValMetrics

			chain := Chain{
				ChainID: "cosmoshub-4",
				Validators: []Validator{
					{ConsAddress: addr},
				},
			}

			tasks := BuildValidatorTasks(&metrics, &client, chain)

			require.Len(t, tasks, 1)
			err := tasks[0].Run(ctx)

			require.NoError(t, err)
			require.Equal(t, client.SigningInfoAddress, addr)

			require.Equal(t, "cosmoshub-4", metrics.GotChain)
			require.Equal(t, addr, metrics.GotAddr)
			require.Equal(t, tt.WantStatus, metrics.GotJailStatus)
		}
	})

	t.Run("happy path - missed blocks", func(t *testing.T) {
		var status SigningInfo
		status.ValSigningInfo.MissedBlocksCounter = "79"

		var client mockValRestClient
		client.StubSigningInfo = status
		client.StubBlock.Block.LastCommit.Height = "1"

		var metrics mockValMetrics
		chain := Chain{
			ChainID: "cosmoshub-4",
			Validators: []Validator{
				{ConsAddress: addr},
			},
		}
		tasks := BuildValidatorTasks(&metrics, &client, chain)

		require.Len(t, tasks, 1)

		err := tasks[0].Run(ctx)
		require.NoError(t, err)

		require.Equal(t, client.SigningInfoAddress, addr)
		require.Equal(t, "cosmoshub-4", metrics.GotChain)
		require.Equal(t, addr, metrics.GotAddr)

		require.Equal(t, float64(79), metrics.GotMissedBlocks)
	})
}
