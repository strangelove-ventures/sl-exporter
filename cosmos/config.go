package cosmos

import "time"

type Chain struct {
	ChainID string
	// Interval is how often to poll the endpoints for data.
	Interval time.Duration
	// Rest are the Cosmos REST (aka LCD) endpoints to poll for data.
	Rest       []Endpoint
	Accounts   []Account
	Validators []Validator
}

type Account struct {
	Address string
	// Alias is a human-readable name for the account, e.g. cosmoshub-validator.
	Alias  string
	Denoms []string
}

type Validator struct {
	// The validator's consensus address. Example prefix: cosmosvalcons...
	ConsAddress string
}

type Endpoint struct {
	URL string
}
