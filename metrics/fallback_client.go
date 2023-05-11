package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type FallbackClient struct {
	hosts   []url.URL
	httpDo  func(req *http.Request) (*http.Response, error)
	metrics ClientMetrics
	rpcType string
}

type ClientMetrics interface {
	IncClientError(rpcType string, host url.URL, errMsg string)
	// TODO(nix): Metrics for request counts. Latency histogram.
}

func NewFallbackClient(client *http.Client, metrics ClientMetrics, rpcType string, hosts []url.URL) *FallbackClient {
	if len(hosts) == 0 {
		panic("no hosts provided")
	}
	return &FallbackClient{
		hosts:   hosts,
		httpDo:  client.Do,
		metrics: metrics,
		rpcType: rpcType,
	}
}

func (c FallbackClient) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	doGet := func(host url.URL) (*http.Response, error) {
		host.Path = path
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, host.String(), nil)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := c.httpDo(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("bad status code %d", resp.StatusCode)
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
