package metrics

type CosmosChain struct {
	// Chain is often the chain id
	Chain string
	// RPCs are the CometBFT RPC servers to query for chain data like block height.
	RPCs []RPC
}

type RPC struct {
	URL string
}
