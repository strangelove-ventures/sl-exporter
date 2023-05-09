package metrics

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/strangelove-ventures/sl-exporter/rest"
)

// CosmosMetrics records metrics for Cosmos chains.
type CosmosMetrics interface {
	SetNodeHeight(chain string, rpcURL url.URL, height float64)
}

// CosmosRestClient queries the Cosmos REST (aka LCD) API.
type CosmosRestClient interface {
	LatestBlock(ctx context.Context, baseURL url.URL) (rest.Block, error)
}

// CosmosRestJob queries the Cosmos REST (aka LCD) API for data and records various metrics.
type CosmosRestJob struct {
	chain    string
	client   CosmosRestClient
	interval time.Duration
	metrics  CosmosMetrics
	url      *url.URL
}

func BuildCosmosRestJobs(metrics CosmosMetrics, client CosmosRestClient, chains []CosmosChain) ([]CosmosRestJob, error) {
	var jobs []CosmosRestJob
	for _, chain := range chains {
		for _, rpc := range chain.Rest {
			u, err := url.Parse(rpc.URL)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, CosmosRestJob{
				chain:    chain.ChainID,
				client:   client,
				interval: rpc.Interval,
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
	if job.interval <= 0 {
		return 5 * time.Second
	}
	return job.interval
}

// Run queries the Endpoint server for data and records various metrics.
func (job CosmosRestJob) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	block, err := job.client.LatestBlock(cctx, *job.url)
	if err != nil {
		return fmt.Errorf("query /status: %w", err)
	}
	height, err := strconv.ParseFloat(block.Block.Header.Height, 64)
	if err != nil {
		return fmt.Errorf("parse height: %w", err)
	}
	job.metrics.SetNodeHeight(job.chain, *job.url, height)
	return nil
}
