package cosmos

import (
	"context"
	"net/url"
	"path"
	"time"
)

// SigningStatus determines whether a validator is jailed or not.
type SigningStatus struct {
	ValSigningInfo struct {
		Address             string    `json:"address"`
		StartHeight         string    `json:"start_height"`
		IndexOffset         string    `json:"index_offset"`
		JailedUntil         time.Time `json:"jailed_until"`
		Tombstoned          bool      `json:"tombstoned"`
		MissedBlocksCounter string    `json:"missed_blocks_counter"`
	} `json:"val_signing_info"`
}

// SigningStatus returns the signing status of a validator given the consensus address.
func (c *RestClient) SigningStatus(ctx context.Context, restURL url.URL, consaddress string) (SigningStatus, error) {
	restURL.Path = path.Join("/cosmos/slashing/v1beta1/signing_infos", consaddress)
	var status SigningStatus
	err := c.get(ctx, restURL.String(), &status)
	return status, err
}
