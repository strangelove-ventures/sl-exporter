package metrics

import "time"

type CosmosChain struct {
	ChainID string
	// RPC are the CometBFT RPC servers to query for chain data like block height.
	RPC []RPC
}

type RPC struct {
	URL string
	// Interval is how often to poll the RPC server for data.
	Interval time.Duration
}
