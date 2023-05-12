package metrics

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// ReferenceAPI records metrics for external http calls.
type ReferenceAPI struct {
	errorCounter *prometheus.CounterVec
	// TODO(nix): Count requests and histogram of latency.
}

func NewHTTPRequest() *ReferenceAPI {
	const subsystem = "reference_api"
	return &ReferenceAPI{
		errorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, subsystem, "error_count"),
				Help: "Number of errors encountered while making external calls to an API to gather reference data.",
			},
			[]string{"host", "reason"},
		),
	}
}

func (c ReferenceAPI) IncAPIError(host url.URL, reason string) {
	c.errorCounter.WithLabelValues(host.Hostname(), reason).Inc()
}

func (c ReferenceAPI) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.errorCounter,
	}
}
