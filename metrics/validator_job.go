package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/strangelove-ventures/sl-exporter/cosmos"
)

type CosmosValidatorMetrics interface {
	SetValJailStatus(chain, consaddress string, status JailStatus)
}

type CosmosValidatorClient interface {
	SigningStatus(ctx context.Context, consaddress string) (cosmos.SigningStatus, error)
}

// CosmosValJob queries the Cosmos REST (aka LCD) API for data and records metrics specific to a validator.
// It records:
// - whether the validator is jailed or tombstoned
type CosmosValJob struct {
	chainID     string
	client      CosmosValidatorClient
	consaddress string
	interval    time.Duration
	metrics     CosmosValidatorMetrics
}

func BuildCosmosValJobs(metrics CosmosValidatorMetrics, client CosmosValidatorClient, chains []CosmosChain) []CosmosValJob {
	var jobs []CosmosValJob
	for _, chain := range chains {
		for _, val := range chain.Validators {
			jobs = append(jobs, CosmosValJob{
				chainID:     chain.ChainID,
				client:      client,
				consaddress: val.ConsAddress,
				interval:    intervalOrDefault(chain.Interval),
				metrics:     metrics,
			})
		}
	}
	return jobs
}

func (job CosmosValJob) String() string {
	return fmt.Sprintf("Cosmos validator %s: %s", job.consaddress, job.chainID)
}

func (job CosmosValJob) Interval() time.Duration { return job.interval }

// Run executes the job gathering a variety of metrics for cosmos validators.
func (job CosmosValJob) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRestTimeout)
	defer cancel()
	resp, err := job.client.SigningStatus(ctx, job.consaddress)
	if err != nil {
		return err
	}
	var status JailStatus
	if time.Since(resp.ValSigningInfo.JailedUntil) < 0 {
		status = JailStatusJailed
	}
	if resp.ValSigningInfo.Tombstoned {
		status = JailStatusTombstoned
	}
	job.metrics.SetValJailStatus(job.chainID, job.consaddress, status)
	return nil
}
