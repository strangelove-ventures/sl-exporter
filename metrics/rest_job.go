package metrics

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/strangelove-ventures/sl-exporter/cosmos"
	"golang.org/x/exp/slog"
)

const (
	defaultInterval    = 15 * time.Second
	defaultRestTimeout = 5 * time.Second
)

func intervalOrDefault(dur time.Duration) time.Duration {
	if dur <= 0 {
		return defaultInterval
	}
	return dur
}

// CosmosMetrics records metrics for Cosmos chains.
type CosmosMetrics interface {
	SetNodeHeight(chain string, height float64)
}

type CosmosRestClient interface {
	LatestBlock(ctx context.Context) (cosmos.Block, error)
}

// CosmosRestJob queries the Cosmos REST (aka LCD) API for data and records various metrics.
type CosmosRestJob struct {
	chainID  string
	client   CosmosRestClient
	interval time.Duration
	metrics  CosmosMetrics
}

func BuildCosmosRestJobs(metrics CosmosMetrics, client CosmosRestClient, chains []CosmosChain) []CosmosRestJob {
	var jobs []CosmosRestJob
	for _, chain := range chains {
		jobs = append(jobs, CosmosRestJob{
			chainID:  chain.ChainID,
			client:   client,
			interval: intervalOrDefault(defaultInterval), // TODO(nix) make configurable
			metrics:  metrics,
		})
	}
	return jobs
}

func (job CosmosRestJob) String() string {
	return fmt.Sprintf("Cosmos REST %s", job.chainID)
}

// Interval is how often to poll the Endpoint server for data. Defaults to 5s.
func (job CosmosRestJob) Interval() time.Duration {
	return intervalOrDefault(job.interval)
}

// Run queries the Endpoint server for data and records various metrics.
func (job CosmosRestJob) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, defaultRestTimeout)
	defer cancel()
	block, err := job.client.LatestBlock(cctx)
	if err != nil {
		return fmt.Errorf("query /status: %w", err)
	}
	if chainID := block.Block.Header.ChainID; chainID != job.chainID {
		slog.Warn("Mismatched chain id", "expected", job.chainID, "actual", chainID)
	}
	height, err := strconv.ParseFloat(block.Block.Header.Height, 64)
	if err != nil {
		return fmt.Errorf("parse height: %w", err)
	}
	job.metrics.SetNodeHeight(job.chainID, height)
	return nil
}
