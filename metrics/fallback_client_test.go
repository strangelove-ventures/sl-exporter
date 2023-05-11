package metrics

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFallbackClient_Get(t *testing.T) {
	urls := []url.URL{
		{Scheme: "http", Host: "1.example.com"},
		{Scheme: "http", Host: "2.example.com"},
	}
	type dummy string // Custom type needed to pass lint
	// Ensures we are passing a unique context.
	ctx := context.WithValue(context.Background(), dummy("test"), dummy("test"))

	t.Run("happy path", func(t *testing.T) {
		client := NewFallbackClient(&http.Client{}, nil, "test", urls)
		require.NotNil(t, client.httpDo)

		var callCount int
		stubResp := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
		client.httpDo = func(req *http.Request) (*http.Response, error) {
			callCount++
			require.Same(t, ctx, req.Context())
			require.Equal(t, "GET", req.Method)
			require.Equal(t, "http://1.example.com/v1/foo", req.URL.String())
			require.Equal(t, "Bar", req.Header.Get("X-Foo"))
			return stubResp, nil
		}

		headers := map[string]string{"X-Foo": "Bar"}
		resp, err := client.Get(ctx, "/v1/foo", headers)
		require.NoError(t, resp.Body.Close())

		require.NoError(t, err)
		require.Same(t, stubResp, resp)
		require.Equal(t, 1, callCount)
	})

	t.Run("fallback on error", func(t *testing.T) {
		client := NewFallbackClient(nil, nil, "test", urls)

		var callCount int
		stubResp := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
		client.httpDo = func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return nil, errors.New("boom")
			}
			require.Same(t, ctx, req.Context())
			require.Equal(t, "GET", req.Method)
			require.Equal(t, "http://2.example.com/v1/foo", req.URL.String())
			require.Equal(t, "Bar", req.Header.Get("X-Foo"))
			return stubResp, nil
		}

		headers := map[string]string{"X-Foo": "Bar"}
		resp, err := client.Get(ctx, "/v1/foo", headers)
		require.NoError(t, resp.Body.Close())

		require.NoError(t, err)
		require.Same(t, stubResp, resp)
		require.Equal(t, 2, callCount)
	})

	t.Run("fallback on bad status code", func(t *testing.T) {
		client := NewFallbackClient(nil, nil, "test", urls)

		var callCount int
		stubResp := &http.Response{StatusCode: http.StatusAccepted, Body: http.NoBody}
		client.httpDo = func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       http.NoBody,
				}, nil
			}
			require.Equal(t, "http://2.example.com", req.URL.String())
			return stubResp, nil
		}

		resp, err := client.Get(ctx, "", nil)
		require.NoError(t, resp.Body.Close())

		require.NoError(t, err)
		require.Same(t, stubResp, resp)
		require.Equal(t, 2, callCount)
	})

	t.Run("all errors", func(t *testing.T) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		client := NewFallbackClient(nil, nil, "test", urls)

		var callCount int
		client.httpDo = func(req *http.Request) (*http.Response, error) {
			callCount++
			switch callCount {
			case 1:
				return nil, errors.New("boom")
			case 2:
				return &http.Response{
					StatusCode: 301 + r.Intn(250),
					Body:       http.NoBody,
				}, nil
			}
			panic("unexpected call count")
		}

		//nolint
		_, err := client.Get(ctx, "", nil)

		require.Error(t, err)
	})
}
