package metrics

import "time"

type CosmosChain struct {
	ChainID string
	// Interval is how often to poll the endpoints for data.
	Interval time.Duration
	// Rest are the Cosmos REST (aka LCD) endpoints to poll for data.
	Rest       []Endpoint
	Validators []CosmosValidator
}

type CosmosValidator struct {
	// The validator's consensus address. Example prefix: cosmosvalcons...
	ConsAddress string
}

type Endpoint struct {
	URL string
}
