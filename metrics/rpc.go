package metrics

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// CosmosMetrics records metrics for Cosmos chains.
type CosmosMetrics interface {
	SetNodeHeight(chain string, rpcURL url.URL, height float64)
}

type RPCClient interface {
	Status(ctx context.Context, rpcURL url.URL) (CometStatus, error)
}

// RPCJob is a job that queries CometBFT (former Tendermint) RPC endpoints for data and records various metrics.
type RPCJob struct {
	chain    string
	client   RPCClient
	interval time.Duration
	metrics  CosmosMetrics
	url      *url.URL
}

func NewRPCJobs(metrics CosmosMetrics, client RPCClient, chains []CosmosChain) ([]RPCJob, error) {
	var jobs []RPCJob
	for _, chain := range chains {
		for _, rpc := range chain.RPCs {
			u, err := url.Parse(rpc.URL)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, RPCJob{
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

func (job RPCJob) String() string {
	return fmt.Sprintf("RPC %s", job.url)
}

// Interval is how often to poll the RPC server for data. Defaults to 5s.
func (job RPCJob) Interval() time.Duration {
	if job.interval <= 0 {
		return 5 * time.Second
	}
	return job.interval
}

// Run queries the RPC server for data and records various metrics.
func (job RPCJob) Run(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	status, err := job.client.Status(cctx, *job.url)
	if err != nil {
		return fmt.Errorf("query /status: %w", err)
	}
	height, err := strconv.ParseFloat(status.Result.SyncInfo.LatestBlockHeight, 64)
	if err != nil {
		return fmt.Errorf("parse height: %w", err)
	}
	job.metrics.SetNodeHeight(job.chain, *job.url, height)
	return nil
}
