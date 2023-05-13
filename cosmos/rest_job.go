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

// RestJob queries the Cosmos REST (aka LCD) API for data and records various metrics.
type RestJob struct {
	chainID  string
	client   Client
	interval time.Duration
	metrics  Metrics
}

func NewRestJob(metrics Metrics, client Client, chain Chain) RestJob {
	return RestJob{
		chainID:  chain.ChainID,
		client:   client,
		interval: intervalOrDefault(chain.Interval),
		metrics:  metrics,
	}
}

func (job RestJob) String() string {
	return fmt.Sprintf("Cosmos REST %s", job.chainID)
}

// Interval is how often to poll the Endpoint server for data. Defaults to 5s.
func (job RestJob) Interval() time.Duration {
	return intervalOrDefault(job.interval)
}

// Run queries the Endpoint server for data and records various metrics.
func (job RestJob) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	block, err := job.client.LatestBlock(cctx)
	if err != nil {
		return err
	}
	if chainID := block.Block.Header.ChainID; chainID != job.chainID {
		slog.Warn("Mismatched chain id", "expected", job.chainID, "actual", chainID, "job", job.String())
	}
	height, err := strconv.ParseFloat(block.Block.Header.Height, 64)
	if err != nil {
		return fmt.Errorf("parse height: %w", err)
	}
	job.metrics.SetNodeHeight(job.chainID, height)
	return nil
}
