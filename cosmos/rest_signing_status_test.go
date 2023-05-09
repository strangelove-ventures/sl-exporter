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

	client := NewRestClient(nil)

	client.httpDo = func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "GET", req.Method)
		require.Equal(t, "https://api.example.com/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons123", req.URL.String())

		const response = `{
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
			Body:       io.NopCloser(strings.NewReader(response)),
		}, nil
	}

	u, err := url.Parse("https://api.example.com")
	require.NoError(t, err)

	got, err := client.SigningStatus(context.Background(), *u, "cosmosvalcons123")
	require.NoError(t, err)

	require.True(t, got.ValSigningInfo.Tombstoned)
	require.Equal(t, time.Date(2021, time.November, 7, 3, 19, 15, 865885008, time.UTC), got.ValSigningInfo.JailedUntil)
}
