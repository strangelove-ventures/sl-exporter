package metrics

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/strangelove-ventures/sl-exporter/cosmos"
	"golang.org/x/exp/slog"
)

func intervalOrDefault(dur time.Duration) time.Duration {
	const defaultInterval = 15 * time.Second
	if dur <= 0 {
		return defaultInterval
	}
	return dur
}

// CosmosMetrics records metrics for Cosmos chains.
type CosmosMetrics interface {
	SetNodeHeight(chain string, rpcURL url.URL, height float64)
}

type CosmosBlockFetcher interface {
	LatestBlock(ctx context.Context, baseURL url.URL) (cosmos.Block, error)
}

// CosmosRestJob queries the Cosmos REST (aka LCD) API for data and records various metrics.
type CosmosRestJob struct {
	chainID  string
	client   CosmosBlockFetcher
	interval time.Duration
	metrics  CosmosMetrics
	url      *url.URL
}

func BuildCosmosRestJobs(metrics CosmosMetrics, client CosmosBlockFetcher, chains []CosmosChain) ([]CosmosRestJob, error) {
	var jobs []CosmosRestJob
	for _, chain := range chains {
		for _, rpc := range chain.Rest {
			u, err := url.Parse(rpc.URL)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, CosmosRestJob{
				chainID:  chain.ChainID,
				client:   client,
				interval: intervalOrDefault(rpc.Interval),
				metrics:  metrics,
				url:      u,
			})
		}
	}
	return jobs, nil
}

func (job CosmosRestJob) String() string {
	return fmt.Sprintf("Cosmos REST %s", job.url)
}

// Interval is how often to poll the Endpoint server for data. Defaults to 5s.
func (job CosmosRestJob) Interval() time.Duration {
	return intervalOrDefault(job.interval)
}

// Run queries the Endpoint server for data and records various metrics.
func (job CosmosRestJob) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	block, err := job.client.LatestBlock(cctx, *job.url)
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
	job.metrics.SetNodeHeight(job.chainID, *job.url, height)
	return nil
}
