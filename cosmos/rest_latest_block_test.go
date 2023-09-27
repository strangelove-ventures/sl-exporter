package cosmos

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/latest_block.json
var latestBlockFixture []byte

type mockHTTPClient struct {
	GetFn func(ctx context.Context, path url.URL) (*http.Response, error)
}

func (m mockHTTPClient) Get(ctx context.Context, path url.URL) (*http.Response, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, path)
	}
	return nil, nil
}

func TestClient_LatestBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		var httpClient mockHTTPClient
		httpClient.GetFn = func(ctx context.Context, path url.URL) (*http.Response, error) {
			require.NotNil(t, ctx)
			require.Equal(t, "/cosmos/base/tendermint/v1beta1/blocks/latest", path.Path)

			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(latestBlockFixture)),
			}, nil
		}
		client := NewRestClient(httpClient)
		got, err := client.LatestBlock(ctx)

		require.NoError(t, err)
		require.Equal(t, "15312655", got.Block.Header.Height)
		require.Equal(t, "15312654", got.Block.LastCommit.Height)
		require.Equal(t, "cosmoshub-4", got.Block.Header.ChainID)
	})

	t.Run("error", func(t *testing.T) {
		var httpClient mockHTTPClient
		httpClient.GetFn = func(ctx context.Context, path url.URL) (*http.Response, error) {
			return nil, errors.New("boom")
		}
		client := NewRestClient(&httpClient)

		_, err := client.LatestBlock(ctx)

		require.Error(t, err)
		require.EqualError(t, err, "boom")
	})
}
