package metrics

import "time"

type CosmosChain struct {
	// Chain is often the chain id
	Chain string
	// RPCs are the CometBFT RPC servers to query for chain data like block height.
	RPCs []RPC
}

type RPC struct {
	URL string
	// Interval is how often to poll the RPC server for data. Defaults to 5s.
	Interval time.Duration
}
