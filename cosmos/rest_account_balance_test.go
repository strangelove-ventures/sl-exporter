package cosmos

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountBalance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		var httpClient mockHTTPClient
		const (
			account = "cosmos123"
			denom   = "ustake"
		)
		httpClient.GetFn = func(ctx context.Context, path string) (*http.Response, error) {
			require.NotNil(t, ctx)
			require.Equal(t, "/cosmos/bank/v1beta1/balances/cosmos123/by_denom?denom=ustake", path)

			const response = `{
  "balance": {
    "denom": "ustake",
    "amount": "1107254710"
  }
}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(response)),
			}, nil
		}
		client := NewRestClient(httpClient)
		got, err := client.AccountBalance(ctx, account, denom)

		require.NoError(t, err)

		require.Equal(t, AccountBalance{
			Account: account,
			Denom:   denom,
			Amount:  1107254710,
		}, got)
	})

	t.Run("error", func(t *testing.T) {
		var httpClient mockHTTPClient
		httpClient.GetFn = func(ctx context.Context, path string) (*http.Response, error) {
			return nil, errors.New("boom")
		}
		client := NewRestClient(&httpClient)

		_, err := client.AccountBalance(ctx, "cosmos123", "ustake")

		require.Error(t, err)
		require.EqualError(t, err, "boom")
	})
}
