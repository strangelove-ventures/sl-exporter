package cosmos

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRestClient_SigningStatus(t *testing.T) {
	t.Parallel()

	var httpClient mockHTTPClient
	httpClient.GetFn = func(ctx context.Context, path url.URL) (*http.Response, error) {
		require.NotNil(t, ctx)
		require.Equal(t, "/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons123", path.Path)

		const fixture = `{
 "val_signing_info": {
   "address": "",
   "start_height": "0",
   "index_offset": "6958718",
   "jailed_until": "2021-11-07T03:19:15.865885008Z",
   "tombstoned": true,
   "missed_blocks_counter": "9"
 }
}`
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(fixture)),
		}, nil
	}
	client := NewRestClient(httpClient)
	got, err := client.SigningInfo(context.Background(), "cosmosvalcons123")
	require.NoError(t, err)

	require.True(t, got.ValSigningInfo.Tombstoned)
	require.Equal(t, time.Date(2021, time.November, 7, 3, 19, 15, 865885008, time.UTC), got.ValSigningInfo.JailedUntil)
}

func TestRestClient_SlashingParams(t *testing.T) {
	t.Parallel()

	var httpClient mockHTTPClient
	httpClient.GetFn = func(ctx context.Context, path url.URL) (*http.Response, error) {
		require.NotNil(t, ctx)
		require.Equal(t, "/cosmos/slashing/v1beta1/params", path.Path)

		const fixture = `{
  "params": {
    "signed_blocks_window": "10000",
    "min_signed_per_window": "0.050000000000000000",
    "downtime_jail_duration": "600s",
    "slash_fraction_double_sign": "0.050000000000000000",
    "slash_fraction_downtime": "0.000100000000000000"
  }
}`
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(fixture)),
		}, nil
	}
	client := NewRestClient(httpClient)
	got, err := client.SlashingParams(context.Background())
	require.NoError(t, err)

	var want SlashingParams
	want.Params.SignedBlocksWindow = "10000"
	want.Params.MinSignedPerWindow = "0.050000000000000000"
	want.Params.DowntimeJailDuration = "600s"
	want.Params.SlashFractionDoubleSign = "0.050000000000000000"
	want.Params.SlashFractionDowntime = "0.000100000000000000"

	require.Equal(t, want, got)
	require.Equal(t, 10000.0, got.SignedBlocksWindow())
}
