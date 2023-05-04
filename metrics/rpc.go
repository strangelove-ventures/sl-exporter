package metrics

import (
	"context"
	"fmt"
	"net/url"
)

type CosmosMetrics interface {
	SetNodeHeight(chain string, rpcURL *url.URL, height float64)
}

type CometFetcher interface {
	Status(ctx context.Context, rpcURL *url.URL) (CometStatus, error)
}

type rpcEndpoint struct {
	chain string
	url   *url.URL
}

// RPCPoller polls CometBFT RPC servers for chain data like block height and records cosmos-specific metrics.
type RPCPoller struct {
	client    CometFetcher
	endpoints []rpcEndpoint
	metrics   CosmosMetrics
}

func NewRPCPoller(metrics CosmosMetrics, client CometFetcher, chains []CosmosChain) (*RPCPoller, error) {
	var endpoints []rpcEndpoint
	for _, chain := range chains {
		for _, rpc := range chain.RPCs {
			u, err := url.Parse(rpc.URL)
			if err != nil {
				return nil, fmt.Errorf("parse url %s: %w", rpc.URL, err)
			}
			endpoints = append(endpoints, rpcEndpoint{
				chain: chain.Chain,
				url:   u,
			})
		}
	}
	return &RPCPoller{
		client:    client,
		endpoints: endpoints,
		metrics:   metrics,
	}, nil
}

func (p *RPCPoller) Poll(ctx context.Context) {
	// listen to context cancel
}
