package cosmos

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/exp/slog"
)

const (
	defaultInterval       = 15 * time.Second
	defaultRequestTimeout = 5 * time.Second
)

func intervalOrDefault(dur time.Duration) time.Duration {
	if dur <= 0 {
		return defaultInterval
	}
	return dur
}

// Metrics records metrics for Cosmos chains.
type Metrics interface {
	SetNodeHeight(chain string, height float64)
}

type Client interface {
	LatestBlock(ctx context.Context) (Block, error)
}

// RestTask queries the Cosmos REST (aka LCD) API for data and records various metrics.
type RestTask struct {
	chainID  string
	client   Client
	interval time.Duration
	metrics  Metrics
}

func (task RestTask) Group() string { return task.chainID }
func (task RestTask) ID() string    { return "latest-block-height" }

func NewRestTask(metrics Metrics, client Client, chain Chain) RestTask {
	return RestTask{
		chainID:  chain.ChainID,
		client:   client,
		interval: intervalOrDefault(chain.Interval),
		metrics:  metrics,
	}
}

// Interval is how often to poll the Endpoint server for data. Defaults to 5s.
func (task RestTask) Interval() time.Duration {
	return intervalOrDefault(task.interval)
}

// Run queries the Endpoint server for data and records various metrics.
func (task RestTask) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	block, err := task.client.LatestBlock(cctx)
	if err != nil {
		return err
	}
	if chainID := block.Block.Header.ChainID; chainID != task.chainID {
		slog.Warn("Mismatched cosmos chain id", "expected", task.chainID, "actual", chainID)
	}
	height, err := strconv.ParseFloat(block.Block.Header.Height, 64)
	if err != nil {
		return fmt.Errorf("parse height: %w", err)
	}
	task.metrics.SetNodeHeight(task.chainID, height)
	return nil
}
