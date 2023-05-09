package rest

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
var LatestBlockFixture []byte

func TestClient_LatestBlock(t *testing.T) {
	// Ensures we aren't comparing against context.Background().
	ctx := context.WithValue(context.Background(), "foo", "bar")

	t.Run("happy path", func(t *testing.T) {
		client := NewClient(http.DefaultClient)
		require.NotNil(t, client.httpDo)

		client.httpDo = func(req *http.Request) (*http.Response, error) {
			require.Same(t, ctx, req.Context())
			require.Equal(t, "GET", req.Method)
			require.Equal(t, "https://api.example.com:443/blocks/latest", req.URL.String())

			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(LatestBlockFixture)),
			}, nil
		}

		u, err := url.Parse("https://api.example.com:443")
		require.NoError(t, err)

		got, err := client.LatestBlock(ctx, *u)
		require.NoError(t, err)

		require.Equal(t, "15226219", got.Block.Header.Height)
	})

	t.Run("http error", func(t *testing.T) {
		client := NewClient(http.DefaultClient)

		client.httpDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("http error")
		}

		_, err := client.LatestBlock(ctx, url.URL{})

		require.Error(t, err)
		require.EqualError(t, err, "http error")
	})

	t.Run("bad status code", func(t *testing.T) {
		client := NewClient(http.DefaultClient)
		require.NotNil(t, client.httpDo)

		client.httpDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Status:     "internal server error",
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}

		_, err := client.LatestBlock(ctx, url.URL{})

		require.Error(t, err)
		require.EqualError(t, err, "internal server error")
	})
}
