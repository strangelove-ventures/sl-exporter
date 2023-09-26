package cosmos

import (
	"context"
	"fmt"
	"time"
)

type AccountMetrics interface {
	SetAccountBalance(chain, alias, address, denom string, balance float64)
}

type AccountClient interface {
	AccountBalance(ctx context.Context, address, denom string) (AccountBalance, error)
}

// AccountTask queries the Cosmos REST (aka LCD) API for account data and records metrics.
type AccountTask struct {
	address  string
	alias    string
	chainID  string
	client   AccountClient
	denom    string
	interval time.Duration
	metrics  AccountMetrics
}

func (task AccountTask) Group() string { return task.chainID }
func (task AccountTask) ID() string    { return fmt.Sprintf("%s-%s", task.address, task.denom) }

func NewAccountTasks(metrics AccountMetrics, client AccountClient, chain Chain) []AccountTask {
	var tasks []AccountTask
	for _, account := range chain.Accounts {
		for _, denom := range account.Denoms {
			tasks = append(tasks, AccountTask{
				address:  account.Address,
				alias:    account.Alias,
				chainID:  chain.ChainID,
				client:   client,
				denom:    denom,
				interval: intervalOrDefault(chain.Interval),
				metrics:  metrics,
			})
		}
	}
	return tasks
}

// Interval is how often to poll the Endpoint server for data. Defaults to 5s.
func (task AccountTask) Interval() time.Duration {
	return task.interval
}

// Run queries the Endpoint server for data and records various metrics.
func (task AccountTask) Run(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	bal, err := task.client.AccountBalance(ctx, task.address, task.denom)
	if err != nil {
		return err
	}
	task.metrics.SetAccountBalance(task.chainID, task.alias, bal.Account, bal.Denom, bal.Amount)
	return nil
}
