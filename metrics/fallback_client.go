package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/exp/slog"
)

type FallbackClient struct {
	hosts   []url.URL
	httpDo  func(req *http.Request) (*http.Response, error)
	log     *slog.Logger
	metrics ClientMetrics
}

type ClientMetrics interface {
	IncAPIError(host url.URL, reason string)
	// TODO(nix): Metrics for request counts. Latency histogram.
}

func NewFallbackClient(client *http.Client, metrics ClientMetrics, hosts []url.URL) *FallbackClient {
	if len(hosts) == 0 {
		panic("no hosts provided")
	}
	return &FallbackClient{
		hosts:   hosts,
		httpDo:  client.Do,
		log:     slog.Default(),
		metrics: metrics,
	}
}

const unknownErrReason = "unknown"

func (c FallbackClient) Get(ctx context.Context, path url.URL) (*http.Response, error) {
	doGet := func(host url.URL) (*http.Response, error) {
		log := c.log.With("host", host.Hostname(), "path", path, "method", http.MethodGet)

		host.Path = path.Path
		host.RawQuery = path.RawQuery

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, host.String(), nil)
		if err != nil {
			log.Debug("Failed to create request", "error", err)
			c.recordErrMetric(host, err)
			return nil, err
		}
		resp, err := c.httpDo(req)
		if err != nil {
			log.Debug("Failed request", "error", err)
			c.recordErrMetric(host, err)
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			log.Debug("Response returned bad status code", "status", resp.StatusCode)
			c.metrics.IncAPIError(host, strconv.Itoa(resp.StatusCode))
			return nil, fmt.Errorf("%s: bad status code %d", req.URL, resp.StatusCode)
		}
		return resp, nil
	}

	var lastErr error
	for _, host := range c.hosts {
		resp, err := doGet(host)
		if err != nil {
			lastErr = err
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

func (c FallbackClient) recordErrMetric(host url.URL, err error) {
	reason := unknownErrReason
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		reason = "timeout"
	case errors.Is(err, context.Canceled):
		// Do not record when the process is shutting down.
		return
	}
	c.metrics.IncAPIError(host, reason)
}
