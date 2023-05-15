package cosmos

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// JailStatus is the status of a validator.
type JailStatus int

const (
	JailStatusActive JailStatus = iota
	JailStatusJailed
	JailStatusTombstoned
)

type ValidatorMetrics interface {
	IncValSignedBlocks(chain, consaddress string)
	SetValJailStatus(chain, consaddress string, status JailStatus)
	SetValSignedBlock(chain, consaddress string, height float64)
	SetValMissedBlocks(chain, consaddress string, missed float64)
}

type ValidatorClient interface {
	LatestBlock(ctx context.Context) (Block, error)
	SigningStatus(ctx context.Context, consaddress string) (SigningStatus, error)
}

// ValidatorJob queries the Cosmos REST (aka LCD) API for data and records metrics specific to a validator.
// It records:
// - whether the validator is jailed or tombstoned
// - the number of blocks signed by the validator
type ValidatorJob struct {
	chainID     string
	client      ValidatorClient
	consaddress string
	interval    time.Duration
	metrics     ValidatorMetrics
}

func BuildValidatorJobs(metrics ValidatorMetrics, client ValidatorClient, chain Chain) []ValidatorJob {
	var jobs []ValidatorJob
	for _, val := range chain.Validators {
		jobs = append(jobs, ValidatorJob{
			chainID:     chain.ChainID,
			client:      client,
			consaddress: val.ConsAddress,
			interval:    intervalOrDefault(chain.Interval),
			metrics:     metrics,
		})
	}
	return jobs
}

func (job ValidatorJob) String() string {
	return fmt.Sprintf("Cosmos validator %s: %s", job.chainID, job.consaddress)
}

func (job ValidatorJob) Interval() time.Duration { return job.interval }

// Run executes the job gathering a variety of metrics for cosmos validators.
func (job ValidatorJob) Run(ctx context.Context) error {
	return errors.Join(
		job.processSigningStatus(ctx),
		job.processSignedBlocks(ctx),
	)
}

func (job ValidatorJob) processSignedBlocks(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	block, err := job.client.LatestBlock(ctx)
	if err != nil {
		return err
	}
	_, valHex, err := bech32.DecodeAndConvert(job.consaddress)
	if err != nil {
		return err
	}
	height, err := strconv.ParseFloat(block.Block.LastCommit.Height, 64)
	if err != nil {
		return fmt.Errorf("parse block last commit height: %w", err)
	}

	for _, sig := range block.Block.LastCommit.Signatures {
		sigHex, err := hex.DecodeString(sig.ValidatorAddress)
		if err != nil {
			return err
		}
		if bytes.Equal(sigHex, valHex) {
			job.metrics.SetValSignedBlock(job.chainID, job.consaddress, height)
			job.metrics.IncValSignedBlocks(job.chainID, job.consaddress)
			break
		}
	}

	return nil
}

func (job ValidatorJob) processSigningStatus(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	resp, err := job.client.SigningStatus(ctx, job.consaddress)
	if err != nil {
		return err
	}

	// Capture jail status
	status := JailStatusActive
	if time.Since(resp.ValSigningInfo.JailedUntil) < 0 {
		status = JailStatusJailed
	}
	if resp.ValSigningInfo.Tombstoned {
		status = JailStatusTombstoned
	}
	job.metrics.SetValJailStatus(job.chainID, job.consaddress, status)

	// Capture missed blocks
	missed, err := strconv.ParseFloat(resp.ValSigningInfo.MissedBlocksCounter, 64)
	if err != nil {
		return fmt.Errorf("parse missed blocks counter: %w", err)
	}
	job.metrics.SetValMissedBlocks(job.chainID, job.consaddress, missed)
	return nil
}
