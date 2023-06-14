package cosmos

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRestClient_SigningStatus(t *testing.T) {
	t.Parallel()

	var httpClient mockHTTPClient
	httpClient.GetFn = func(ctx context.Context, path string) (*http.Response, error) {
		require.NotNil(t, ctx)
		require.Equal(t, "/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons123", path)

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
