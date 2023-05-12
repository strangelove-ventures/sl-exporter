package metrics

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// ReferenceRPC records metrics for external RPC calls.
type ReferenceRPC struct {
	errorCounter *prometheus.CounterVec
	// TODO(nix): Count requests and histogram of latency.
}

func NewReferenceRPC() *ReferenceRPC {
	const subsystem = "reference_rpc"
	return &ReferenceRPC{
		errorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(Namespace, subsystem, "error_count"),
				Help: "Number of errors encountered while making external RPC, API, or GRPC calls.",
			},
			[]string{"type", "host", "reason"},
		),
	}
}

func (c ReferenceRPC) IncClientError(rpcType string, host url.URL, reason string) {
	c.errorCounter.WithLabelValues(rpcType, host.Hostname(), reason).Inc()
}

func (c ReferenceRPC) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.errorCounter,
	}
}
