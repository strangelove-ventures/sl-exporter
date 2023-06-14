package cosmos

import (
	"context"
	"time"
)

type ParamsClient interface {
	SlashingParams(ctx context.Context) (SlashingParams, error)
}

type ParamsMetrics interface {
	SetValSlashingParams(chain string, window float64)
}

type ParamsTask struct {
	chainID string
	client  ParamsClient
	metrics ParamsMetrics
}

func NewParamsTask(metrics ParamsMetrics, client ParamsClient, chain Chain) ParamsTask {
	return ParamsTask{
		chainID: chain.ChainID,
		client:  client,
		metrics: metrics,
	}
}

func (p ParamsTask) Group() string { return p.chainID }
func (p ParamsTask) ID() string    { return "params" }

// Interval is hardcoded to a longer duration because params change rarely.
// They require a gov proposal. Additionally, longer duration minimizes API calls to prevent hitting rate limits.
func (p ParamsTask) Interval() time.Duration { return 5 * time.Minute }

func (p ParamsTask) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	slash, err := p.client.SlashingParams(cctx)
	if err != nil {
		return err
	}
	p.metrics.SetValSlashingParams(p.chainID, slash.SignedBlocksWindow())
	return nil
}
