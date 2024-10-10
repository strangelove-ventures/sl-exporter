package metrics

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

type mockClientMetrics struct {
	IncClientErrCalls int
	GotHost           url.URL
	GotErrMsg         string
}

func (m *mockClientMetrics) IncAPIError(host url.URL, errMsg string) {
	m.IncClientErrCalls++
	m.GotHost = host
	m.GotErrMsg = errMsg
}

var nopLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

func TestFallbackClient_Get(t *testing.T) {
	urls := []url.URL{
		{Scheme: "http", Host: "1.example.com"},
		{Scheme: "http", Host: "2.example.com"},
	}
	type dummy string // Custom type needed to pass lint
	// Ensures we are passing a unique context.
	ctx := context.WithValue(context.Background(), dummy("test"), dummy("test"))

	t.Run("happy path", func(t *testing.T) {
		client := NewFallbackClient(&http.Client{}, nil, urls)
		client.log = nopLogger
		require.NotNil(t, client.httpDo)

		var callCount int
		stubResp := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
		client.httpDo = func(req *http.Request) (*http.Response, error) {
			callCount++
			require.Same(t, ctx, req.Context())
			require.Equal(t, "GET", req.Method)
			require.Equal(t, "http://1.example.com/v1/foo", req.URL.String())
			return stubResp, nil
		}

		resp, err := client.Get(ctx, url.URL{Path: "/v1/foo"})
		require.NoError(t, resp.Body.Close())

		require.NoError(t, err)
		require.Same(t, stubResp, resp)
		require.Equal(t, 1, callCount)
	})

	t.Run("fallback on error", func(t *testing.T) {
		var metrics mockClientMetrics
		client := NewFallbackClient(nil, &metrics, urls)
		client.log = nopLogger

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
			return stubResp, nil
		}

		resp, err := client.Get(ctx, url.URL{Path: "/v1/foo"})
		require.NoError(t, resp.Body.Close())

		require.NoError(t, err)
		require.Same(t, stubResp, resp)
		require.Equal(t, 2, callCount)
	})

	t.Run("fallback on bad status code", func(t *testing.T) {
		var metrics mockClientMetrics
		client := NewFallbackClient(nil, &metrics, urls)
		client.log = nopLogger

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

		resp, err := client.Get(ctx, url.URL{})
		require.NoError(t, resp.Body.Close())

		require.NoError(t, err)
		require.Same(t, stubResp, resp)
		require.Equal(t, 2, callCount)
	})

	t.Run("all errors", func(t *testing.T) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		var metrics mockClientMetrics
		client := NewFallbackClient(nil, &metrics, urls)
		client.log = nopLogger

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
		_, err := client.Get(ctx, url.URL{})

		require.Error(t, err)
	})

	t.Run("error metrics", func(t *testing.T) {
		for _, tt := range []struct {
			Err      error
			Response *http.Response
			WantMsg  string
		}{
			{errors.New("boom"), nil, "unknown"},
			{fmt.Errorf("deadline: %w", context.DeadlineExceeded), nil, "timeout"},
			{nil, &http.Response{StatusCode: http.StatusNotFound}, "404"},
		} {
			var metrics mockClientMetrics
			client := NewFallbackClient(nil, &metrics, []url.URL{{Host: "error.example.com"}})
			client.log = nopLogger

			client.httpDo = func(req *http.Request) (*http.Response, error) {
				if tt.Response != nil {
					tt.Response.Body = http.NoBody
				}
				return tt.Response, tt.Err
			}

			//nolint
			_, _ = client.Get(ctx, url.URL{})

			require.Equal(t, "error.example.com", metrics.GotHost.Hostname(), tt)
			require.Equal(t, tt.WantMsg, metrics.GotErrMsg, tt)
		}
	})

	t.Run("context canceled error", func(t *testing.T) {
		var metrics mockClientMetrics
		client := NewFallbackClient(nil, &metrics, []url.URL{{Host: "error.example.com"}})
		client.log = nopLogger

		client.httpDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.Join(context.Canceled)
		}

		//nolint
		_, _ = client.Get(ctx, url.URL{})

		require.Zero(t, metrics.IncClientErrCalls)
	})
}
