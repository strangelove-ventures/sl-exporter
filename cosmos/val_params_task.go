package cosmos

import (
	"context"
	"time"
)

type ValParamsClient interface {
	SlashingParams(ctx context.Context) (SlashingParams, error)
}

type ValParamsMetrics interface {
	SetValSlashingParams(chain string, window float64)
}

type ValParamsTask struct {
	chainID string
	client  ValParamsClient
	metrics ValParamsMetrics
}

func NewValParamsTask(metrics ValParamsMetrics, client ValParamsClient, chain Chain) ValParamsTask {
	return ValParamsTask{
		chainID: chain.ChainID,
		client:  client,
		metrics: metrics,
	}
}

func (p ValParamsTask) Group() string { return p.chainID }
func (p ValParamsTask) ID() string    { return "params" }

// Interval is hardcoded to a longer duration because params rarely change.
// They require a gov proposal. Additionally, longer duration minimizes API calls to prevent hitting rate limits.
func (p ValParamsTask) Interval() time.Duration { return 5 * time.Minute }

func (p ValParamsTask) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	slash, err := p.client.SlashingParams(cctx)
	if err != nil {
		return err
	}
	p.metrics.SetValSlashingParams(p.chainID, slash.SignedBlocksWindow())
	return nil
}
