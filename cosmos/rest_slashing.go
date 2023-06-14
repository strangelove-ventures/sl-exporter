package cosmos

import (
	"context"
	"path"
	"time"
)

// SigningInfo determines whether a validator is jailed or not.
type SigningInfo struct {
	ValSigningInfo struct {
		Address             string    `json:"address"`
		StartHeight         string    `json:"start_height"`
		IndexOffset         string    `json:"index_offset"`
		JailedUntil         time.Time `json:"jailed_until"`
		Tombstoned          bool      `json:"tombstoned"`
		MissedBlocksCounter string    `json:"missed_blocks_counter"`
	} `json:"val_signing_info"`
}

// SigningInfo returns the signing status of a validator given the consensus address.
// Docs: https://docs.cosmos.network/swagger/#/Query/SigningInfo
func (c RestClient) SigningInfo(ctx context.Context, consaddress string) (SigningInfo, error) {
	p := path.Join("/cosmos/slashing/v1beta1/signing_infos", consaddress)
	var status SigningInfo
	err := c.get(ctx, p, &status)
	return status, err
}
