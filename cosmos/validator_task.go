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
	var tasks []ValidatorJob
	for _, val := range chain.Validators {
		tasks = append(tasks, ValidatorJob{
			chainID:     chain.ChainID,
			client:      client,
			consaddress: val.ConsAddress,
			interval:    intervalOrDefault(chain.Interval),
			metrics:     metrics,
		})
	}
	return tasks
}

func (task ValidatorJob) String() string {
	return fmt.Sprintf("Cosmos validator %s: %s", task.chainID, task.consaddress)
}

func (task ValidatorJob) Interval() time.Duration { return task.interval }

// Run executes the job gathering a variety of metrics for cosmos validators.
func (task ValidatorJob) Run(ctx context.Context) error {
	return errors.Join(
		task.processSigningStatus(ctx),
		task.processSignedBlocks(ctx),
	)
}

func (task ValidatorJob) processSignedBlocks(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	block, err := task.client.LatestBlock(ctx)
	if err != nil {
		return err
	}
	_, valHex, err := bech32.DecodeAndConvert(task.consaddress)
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
			task.metrics.SetValSignedBlock(task.chainID, task.consaddress, height)
			task.metrics.IncValSignedBlocks(task.chainID, task.consaddress)
			break
		}
	}

	return nil
}

func (task ValidatorJob) processSigningStatus(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	resp, err := task.client.SigningStatus(ctx, task.consaddress)
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
	task.metrics.SetValJailStatus(task.chainID, task.consaddress, status)

	// Capture missed blocks
	missed, err := strconv.ParseFloat(resp.ValSigningInfo.MissedBlocksCounter, 64)
	if err != nil {
		return fmt.Errorf("parse missed blocks counter: %w", err)
	}
	task.metrics.SetValMissedBlocks(task.chainID, task.consaddress, missed)
	return nil
}
