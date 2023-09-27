package cosmos

import (
	"context"
	"net/url"
	"path"
	"strconv"
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
	var info SigningInfo
	err := c.get(ctx, url.URL{Path: p}, &info)
	return info, err
}

type SlashingParams struct {
	Params struct {
		SignedBlocksWindow      string `json:"signed_blocks_window"`
		MinSignedPerWindow      string `json:"min_signed_per_window"`
		DowntimeJailDuration    string `json:"downtime_jail_duration"`
		SlashFractionDoubleSign string `json:"slash_fraction_double_sign"`
		SlashFractionDowntime   string `json:"slash_fraction_downtime"`
	} `json:"params"`
}

func (s SlashingParams) SignedBlocksWindow() float64 {
	v, _ := strconv.ParseFloat(s.Params.SignedBlocksWindow, 64)
	return v
}

// SlashingParams returns the slashing parameters.
// Docs: https://docs.cosmos.network/swagger/#/Query/SlashingParams
func (c RestClient) SlashingParams(ctx context.Context) (SlashingParams, error) {
	var params SlashingParams
	err := c.get(ctx, url.URL{Path: "/cosmos/slashing/v1beta1/params"}, &params)
	return params, err
}
