package metrics

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/strangelove-ventures/sl-exporter/cosmos"
)

type CosmosValidatorMetrics interface {
	SetValJailStatus(chain, consaddress string, restURL url.URL, status JailStatus)
}

type CosmosValidatorClient interface {
	SigningStatus(ctx context.Context, baseURL url.URL, consaddress string) (cosmos.SigningStatus, error)
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
	restURL     *url.URL
}

func BuildCosmosValJobs(metrics CosmosValidatorMetrics, client CosmosValidatorClient, chains []CosmosChain) ([]CosmosValJob, error) {
	var jobs []CosmosValJob
	for _, chain := range chains {
		for _, val := range chain.Validators {
			for _, rpc := range chain.Rest {
				u, err := url.Parse(rpc.URL)
				if err != nil {
					return nil, fmt.Errorf("parse url: %w", err)
				}
				jobs = append(jobs, CosmosValJob{
					chainID:     chain.ChainID,
					client:      client,
					consaddress: val.ConsAddress,
					interval:    intervalOrDefault(rpc.Interval),
					metrics:     metrics,
					restURL:     u,
				})
			}
		}
	}
	return jobs, nil
}

func (job CosmosValJob) String() string {
	return fmt.Sprintf("Cosmos validator %s: %s", job.consaddress, job.restURL)
}

func (job CosmosValJob) Interval() time.Duration { return job.interval }

func (job CosmosValJob) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRestTimeout)
	defer cancel()
	resp, err := job.client.SigningStatus(ctx, *job.restURL, job.consaddress)
	if err != nil {
		return fmt.Errorf("signing status: %w", err)
	}
	var status JailStatus
	if time.Since(resp.ValSigningInfo.JailedUntil) < 0 {
		status = JailStatusJailed
	}
	if resp.ValSigningInfo.Tombstoned {
		status = JailStatusTombstoned
	}
	job.metrics.SetValJailStatus(job.chainID, job.consaddress, *job.restURL, status)
	return nil
}
